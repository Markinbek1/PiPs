package autostart

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	regKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName    = "OperaPiPToggle"
)

// IsEnabled checks if the application is set to launch on startup
func IsEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, regKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetStringValue(appName)
	if err != nil {
		return false
	}

	execPath, err := os.Executable()
	if err != nil {
		return val != ""
	}

	cleanVal := strings.Trim(val, `"`)
	return strings.EqualFold(filepath.Clean(cleanVal), filepath.Clean(execPath))
}

// SetEnabled enables or disables startup registration in HKCU
func SetEnabled(enable bool) error {
	var k registry.Key
	var err error
	if enable {
		k, _, err = registry.CreateKey(registry.CURRENT_USER, regKeyPath, registry.SET_VALUE)
	} else {
		k, err = registry.OpenKey(registry.CURRENT_USER, regKeyPath, registry.SET_VALUE)
		if err == registry.ErrNotExist {
			return nil
		}
	}
	if err != nil {
		return err
	}
	defer k.Close()

	if enable {
		execPath, err := os.Executable()
		if err != nil {
			return err
		}
		formattedPath := `"` + execPath + `"`
		err = k.SetStringValue(appName, formattedPath)
		if err != nil {
			log.Printf("[autostart] Error setting registry value: %v", err)
			return err
		}
		log.Printf("[autostart] Successfully enabled autostart for %s", formattedPath)
	} else {
		err = k.DeleteValue(appName)
		if err != nil && err != registry.ErrNotExist {
			log.Printf("[autostart] Error deleting registry value: %v", err)
			return err
		}
		log.Println("[autostart] Successfully disabled autostart")
	}
	return nil
}
