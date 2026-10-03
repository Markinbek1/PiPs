package winapi

import (
	"path/filepath"
	"strings"
)

// IsSupportedBrowserProcess limits window changes to the supported browsers.
func IsSupportedBrowserProcess(processName string) bool {
	switch strings.ToLower(filepath.Base(processName)) {
	case "opera.exe", "chrome.exe", "msedge.exe":
		return true
	default:
		return false
	}
}

// IsPiPTitle accepts complete native PiP titles, never a substring in a tab title.
func IsPiPTitle(title string) bool {
	title = strings.NewReplacer("-", " ", "\u2011", " ", "\u00a0", " ").Replace(title)
	title = strings.ToLower(strings.Join(strings.Fields(title), " "))
	switch title {
	case "picture in picture", "картинка в картинке", "pip":
		return true
	default:
		return false
	}
}

func hasPiPWindowStyle(className string, exStyle uintptr) bool {
	return className == "Chrome_WidgetWin_1" && exStyle&WS_EX_TOPMOST != 0
}
