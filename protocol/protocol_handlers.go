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

package protocol

import (
	"net"

	"github.com/ivpn/desktop-app-daemon/protocol/types"
	"github.com/ivpn/desktop-app-daemon/service/preferences"
	"github.com/ivpn/desktop-app-daemon/version"
)

// OnServiceSessionChanged - SessionChanged handler
func (p *Protocol) OnServiceSessionChanged() {
	service := p._service
	if service == nil {
		return
	}

	// send back Hello message with account session info
	helloResp := types.HelloResp{
		Version: version.Version(),
		Session: types.CreateSessionResp(service.Preferences().Session)}

	p.notifyClients(&helloResp)
}

// OnAccountStatus - handler of account status info. Notifying clients.
func (p *Protocol) OnAccountStatus(sessionToken string, accountInfo preferences.AccountStatus) {
	if len(sessionToken) == 0 {
		return
	}

	p.notifyClients(&types.AccountStatusResp{
		SessionToken: sessionToken,
		Account:      accountInfo})
}

// OnDNSChanged - DNS changed handler
func (p *Protocol) OnDNSChanged(dns net.IP) {
	// notify all clients
	if dns == nil {
		p.notifyClients(&types.SetAlternateDNSResp{IsSuccess: true, ChangedDNS: ""})
	} else {
		p.notifyClients(&types.SetAlternateDNSResp{IsSuccess: true, ChangedDNS: dns.String()})
	}
}

// OnKillSwitchStateChanged - Firewall change handler
func (p *Protocol) OnKillSwitchStateChanged() {
	// notify all clients about KillSwitch status
	if isEnabled, isPersistant, isAllowLAN, isAllowLanMulticast, err := p._service.KillSwitchState(); err != nil {
		log.Error(err)
	} else {
		p.notifyClients(&types.KillSwitchStatusResp{IsEnabled: isEnabled, IsPersistent: isPersistant, IsAllowLAN: isAllowLAN, IsAllowMulticast: isAllowLanMulticast})
	}
}
