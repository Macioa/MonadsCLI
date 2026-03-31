package types

import (
	"runtime"
	"strings"
)

// CLI defines metadata for a command line interface integration.
type CLI struct {
	Name                 string `json:"name"`
	KeyURL               string `json:"keyUrl"`
	KeyENV               string `json:"keyEnv"`
	Codename             string `json:"codename"` // Uppercase alias for settings (e.g. CURSOR); used when set, else Command
	Command              string `json:"command"`
	Prompt               string `json:"prompt"`
	LinuxInstall         string `json:"linuxInstall"`
	WindowsInstallString string `json:"windowsInstallString"`
}

// InstallCommand returns the install script for the current OS.
func (c CLI) InstallCommand() string {
	if runtime.GOOS == "windows" && strings.TrimSpace(c.WindowsInstallString) != "" {
		return c.WindowsInstallString
	}
	return c.LinuxInstall
}
