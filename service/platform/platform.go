package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ivpn/desktop-app-daemon/service/platform/filerights"
)

var (
	settingsFile    string
	servicePortFile string
	serversFile     string
	logFile         string
	openvpnLogFile  string

	openVpnBinaryPath    string
	openvpnCaKeyFile     string
	openvpnTaKeyFile     string
	openvpnConfigFile    string
	openvpnUpScript      string
	openvpnDownScript    string
	openvpnProxyAuthFile string

	obfsproxyStartScript string
	obfsproxyHostPort    int

	wgBinaryPath     string
	wgToolBinaryPath string
	wgConfigFilePath string
)

func init() {
	// initialize all constant values (e.g. servicePortFile) which can be used in external projects (IVPN CLI)
	doInitConstants()
	if len(servicePortFile) <= 0 {
		panic("Path to service port file not defined ('platform.servicePortFile' is empty)")
	}
}

// Init - initialize all preferences required for a daemon
// Must be called on beginning of application start by a daemon(service)
func Init() (warnings []string, errors []error) {

	obfsproxyHostPort = 5145

	// do variables initialization for current OS
	warnings, errors = doOsInit()
	if errors == nil {
		errors = make([]error, 0)
	}
	if warnings == nil {
		warnings = make([]string, 0)
	}

	// creating required folders
	if err := makeDir("servicePortFile", filepath.Dir(servicePortFile)); err != nil {
		errors = append(errors, err)
	}
	if err := makeDir("logFile", filepath.Dir(logFile)); err != nil {
		errors = append(errors, err)
	}
	if err := makeDir("openvpnLogFile", filepath.Dir(openvpnLogFile)); err != nil {
		errors = append(errors, err)
	}
	if err := makeDir("settingsFile", filepath.Dir(settingsFile)); err != nil {
		errors = append(errors, err)
	}
	if err := makeDir("openvpnConfigFile", filepath.Dir(openvpnConfigFile)); err != nil {
		errors = append(errors, err)
	}
	if err := makeDir("wgConfigFilePath", filepath.Dir(wgConfigFilePath)); err != nil {
		errors = append(errors, err)
	}

	// checking file permissions
	if err := checkFileAccessRigthsStaticConfig("openvpnCaKeyFile", openvpnCaKeyFile); err != nil {
		errors = append(errors, err)
	}
	if err := checkFileAccessRigthsStaticConfig("openvpnTaKeyFile", openvpnTaKeyFile); err != nil {
		errors = append(errors, err)
	}

	if len(openvpnUpScript) > 0 {
		if err := checkFileAccessRigthsExecutable("openvpnUpScript", openvpnUpScript); err != nil {
			errors = append(errors, err)
		}
	}

	if len(openvpnDownScript) > 0 {
		if err := checkFileAccessRigthsExecutable("openvpnDownScript", openvpnDownScript); err != nil {
			errors = append(errors, err)
		}
	}

	// checking availability of OpenVPN binaries
	if err := checkFileAccessRigthsExecutable("openVpnBinaryPath", openVpnBinaryPath); err != nil {
		warnings = append(warnings, fmt.Errorf("OpenVPN functionality not accessible: %w", err).Error())
	}
	// checking availability of obfsproxy binaries
	if err := checkFileAccessRigthsExecutable("obfsproxyStartScript", obfsproxyStartScript); err != nil {
		warnings = append(warnings, fmt.Errorf("obfsproxy functionality not accessible: %w", err).Error())
	}
	// checling availability of WireGuard binaries
	if err := checkFileAccessRigthsExecutable("wgBinaryPath", wgBinaryPath); err != nil {
		warnings = append(warnings, fmt.Errorf("WireGuard functionality not accessible: %w", err).Error())
	}
	if err := checkFileAccessRigthsExecutable("wgToolBinaryPath", wgToolBinaryPath); err != nil {
		warnings = append(warnings, fmt.Errorf("WireGuard functionality not accessible: %w", err).Error())
	}

	w, e := doInitOperations()
	if len(w) > 0 {
		warnings = append(warnings, w)
	}
	if e != nil {
		errors = append(errors, e)
	}

	return warnings, errors
}

func checkFileAccessRigthsStaticConfig(paramName string, file string) error {
	if err := filerights.CheckFileAccessRigthsStaticConfig(file); err != nil {
		return fmt.Errorf("(%s) %w", paramName, err)
	}
	return nil
}

func checkFileAccessRigthsExecutable(paramName string, file string) error {
	if err := filerights.CheckFileAccessRigthsExecutable(file); err != nil {
		return fmt.Errorf("(%s) %w", paramName, err)
	}
	return nil
}

func makeDir(description string, dirpath string) error {
	if len(dirpath) == 0 {
		return fmt.Errorf("parameter not initialized: %s", description)
	}

	if err := os.MkdirAll(dirpath, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create directory error: %s (%s:%s)", err.Error(), description, dirpath)
	}
	return nil
}

// Is64Bit - returns 'true' if binary compiled in 64-bit architecture
func Is64Bit() bool {
	if strconv.IntSize == 64 {
		return true
	}
	return false
}

// SettingsFile path to settings file
func SettingsFile() string {
	return settingsFile
}

// ServicePortFile parh to service port file
func ServicePortFile() string {
	return servicePortFile
}

// ServersFile path to servers.json
func ServersFile() string {
	return serversFile
}

// LogFile path to log-file
func LogFile() string {
	return logFile
}

// OpenvpnLogFile path to log-file for openvpn
func OpenvpnLogFile() string {
	return openvpnLogFile
}

// OpenVpnBinaryPath path to openvpn binary
func OpenVpnBinaryPath() string {
	return openVpnBinaryPath
}

// OpenvpnCaKeyFile path to openvpn CA key file
func OpenvpnCaKeyFile() string {
	return openvpnCaKeyFile
}

// OpenvpnTaKeyFile path to openvpn TA key file
func OpenvpnTaKeyFile() string {
	return openvpnTaKeyFile
}

// OpenvpnConfigFile path to openvpn config file
func OpenvpnConfigFile() string {
	return openvpnConfigFile
}

// OpenvpnUpScript path to openvpn UP script file
func OpenvpnUpScript() string {
	return openvpnUpScript
}

// OpenvpnDownScript path to openvpn Down script file
func OpenvpnDownScript() string {
	return openvpnDownScript
}

// OpenvpnProxyAuthFile path to openvpn proxy credentials file
func OpenvpnProxyAuthFile() string {
	return openvpnProxyAuthFile
}

// ObfsproxyStartScript path to obfsproxy binary
func ObfsproxyStartScript() string {
	return obfsproxyStartScript
}

// ObfsproxyHostPort is an port of obfsproxy host
func ObfsproxyHostPort() int {
	return obfsproxyHostPort
}

// WgBinaryPath path to WireGuard binary
func WgBinaryPath() string {
	return wgBinaryPath
}

// WgToolBinaryPath path to WireGuard tools binary
func WgToolBinaryPath() string {
	return wgToolBinaryPath
}

// WGConfigFilePath path to WireGuard configuration file
func WGConfigFilePath() string {
	return wgConfigFilePath
}
