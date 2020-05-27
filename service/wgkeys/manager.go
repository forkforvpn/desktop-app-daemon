//
//  Daemon for IVPN Client Desktop
//  https://github.com/ivpn/desktop-app-daemon
//
//  Created by Stelnykovych Alexandr.
//  Copyright (c) 2020 Privatus Limited.
//
//  This file is part of the Daemon for IVPN Client Desktop.
//
//  The Daemon for IVPN Client Desktop is free software: you can redistribute it and/or
//  modify it under the terms of the GNU General Public License as published by the Free
//  Software Foundation, either version 3 of the License, or (at your option) any later version.
//
//  The Daemon for IVPN Client Desktop is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY
//  or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
//  details.
//
//  You should have received a copy of the GNU General Public License
//  along with the Daemon for IVPN Client Desktop. If not, see <https://www.gnu.org/licenses/>.
//

package wgkeys

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ivpn/desktop-app-daemon/api"
	"github.com/ivpn/desktop-app-daemon/logger"
	"github.com/ivpn/desktop-app-daemon/vpn/wireguard"
)

var log *logger.Logger

func init() {
	log = logger.NewLogger("wgkeys")
}

//HardExpirationIntervalDays = 40;

// IWgKeysChangeReceiver WG key update handler
type IWgKeysChangeReceiver interface {
	WireGuardSaveNewKeys(wgPublicKey string, wgPrivateKey string, wgLocalIP net.IP)
	WireGuardGetKeys() (session, wgPublicKey, wgPrivateKey, wgLocalIP string, generatedTime time.Time, updateInterval time.Duration)
	Connected() bool
}

// CreateKeysManager create WireGuard keys manager
func CreateKeysManager(apiObj *api.API, wgToolBinPath string) *KeysManager {
	return &KeysManager{
		_stopKeysRotation: make(chan struct{}),
		_wgToolBinPath:    wgToolBinPath,
		_apiObj:           apiObj}
}

// KeysManager WireGuard keys manager
type KeysManager struct {
	_mutex            sync.Mutex
	_service          IWgKeysChangeReceiver
	_apiObj           *api.API
	_wgToolBinPath    string
	_stopKeysRotation chan struct{}
}

// Init - initialize master service
func (m *KeysManager) Init(receiver IWgKeysChangeReceiver) error {
	if receiver == nil || m._service != nil {
		return fmt.Errorf("failed to initialize WG KeysManager")
	}
	m._service = receiver
	return nil
}

// StartKeysRotation start keys rotation
func (m *KeysManager) StartKeysRotation() error {
	if m._service == nil {
		return fmt.Errorf("unable to start WG keys rotation (KeysManager not initialized)")
	}

	m.StopKeysRotation()

	_, activePublicKey, _, _, lastUpdate, interval := m._service.WireGuardGetKeys()
	if interval <= 0 {
		return fmt.Errorf("unable to start WG keys rotation (update interval not defined)")
	}

	if len(activePublicKey) == 0 {
		log.Info("Active public WG key is not defined. WG key rotation disabled.")
		return nil
	}

	go func() {
		log.Info(fmt.Sprintf("Keys rotation started (interval:%v)", interval))
		defer log.Info("Keys rotation stopped")

		needStop := false
		isLastUpdateFailed := false

		for needStop == false {
			_, _, _, _, lastUpdate, interval = m._service.WireGuardGetKeys()
			waitInterval := time.Until(lastUpdate.Add(interval))
			if isLastUpdateFailed {
				waitInterval = time.Hour
				lastUpdate = time.Now()
			}

			// update immediately, if it is a time
			if lastUpdate.Add(waitInterval).Before(time.Now()) {
				waitInterval = time.Second
			}

			select {
			case <-time.After(waitInterval):
				err := m.UpdateKeysIfNecessary()
				if err != nil {
					isLastUpdateFailed = true
				} else {
					isLastUpdateFailed = false
					lastUpdate = time.Now()
				}

				break

			case <-m._stopKeysRotation:
				needStop = true
				break
			}
		}
	}()

	return nil
}

// StopKeysRotation stop keys rotation
func (m *KeysManager) StopKeysRotation() {
	select {
	case m._stopKeysRotation <- struct{}{}:
	default:
	}
}

// GenerateKeys generate keys
func (m *KeysManager) GenerateKeys() error {
	return m.generateKeys(false)
}

// UpdateKeysIfNecessary generate or update keys
// 1) If no active WG keys defined - new keys will be generated + key rotation will be started
// 2) If active WG key defined - key will be updated only if it is a time to do it
func (m *KeysManager) UpdateKeysIfNecessary() error {
	return m.generateKeys(true)
}

func (m *KeysManager) generateKeys(onlyUpdateIfNecessary bool) (retErr error) {
	defer func() {
		if retErr != nil {
			log.Error("Failed to update WG keys: ", retErr)
		}
	}()

	if m._service == nil {
		return fmt.Errorf("WG KeysManager not initialized")
	}

	// Check update configuration
	// (not blocked by mutex because in order to return immediately if nothing to do)
	session, activePublicKey, _, _, lastUpdate, interval := m._service.WireGuardGetKeys()

	// function to check if update required
	isNecessaryUpdate := func() (bool, error) {
		if onlyUpdateIfNecessary == false {
			return true, nil
		}
		if interval <= 0 {
			// update interval must be defined
			return false, fmt.Errorf("unable to 'GenerateOrUpdateKeys' (update interval is not defined)")
		}
		if len(activePublicKey) > 0 {
			// If active WG key defined - key will be updated only if it is a time to do it
			if lastUpdate.Add(interval).After(time.Now()) {
				// it is not a time to regenerate keys
				return false, nil
			}
		}
		return true, nil
	}

	if haveToUpdate, err := isNecessaryUpdate(); haveToUpdate == false || err != nil {
		return err
	}

	m._mutex.Lock()
	defer m._mutex.Unlock()

	// Check update configuration second time (locked by mutex)
	session, activePublicKey, _, _, lastUpdate, interval = m._service.WireGuardGetKeys()
	if haveToUpdate, err := isNecessaryUpdate(); haveToUpdate == false || err != nil {
		return err
	}

	log.Info("Updating WG keys...")

	pub, priv, err := wireguard.GenerateKeys(m._wgToolBinPath)
	if err != nil {
		return err
	}

	activeKeyToUpdate := activePublicKey
	// When VPN is not connected - no sense to use 'update',
	// just set new WG key for this session.
	// This can avoid any potential issues regarding 'WgPublicKeyNotFound' error.
	if m._service.Connected() == false {
		activeKeyToUpdate = ""
	}

	localIP, err := m._apiObj.WireGuardKeySet(session, pub, activeKeyToUpdate)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("WG keys updated (%s:%s) ", localIP.String(), pub))

	// notify service about new keys
	m._service.WireGuardSaveNewKeys(pub, priv, localIP)

	// If no active WG keys defined - new keys will be generated + key rotation will be started
	if len(activePublicKey) == 0 {
		m.StartKeysRotation()
	}

	return nil
}
