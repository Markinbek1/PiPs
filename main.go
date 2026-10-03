package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/getlantern/systray"

	"pip-toggle/pkg/autostart"
	"pip-toggle/pkg/hotkey"
	"pip-toggle/pkg/i18n"
	"pip-toggle/pkg/icon"
	"pip-toggle/pkg/notify"
	"pip-toggle/pkg/winapi"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutexW        = kernel32.NewProc("CreateMutexW")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")

	currentStatus        bool
	mStatus              *systray.MenuItem
	uiLanguage           = i18n.Russian
	languageSettingsPath string
)

func tr(key i18n.Key) string {
	return i18n.Text(uiLanguage, key)
}

const (
	ERROR_ALREADY_EXISTS = 183
	WAIT_TIMEOUT         = 258
)

func checkSingleInstance() (uintptr, error) {
	mutexName, err := syscall.UTF16PtrFromString("Local\\OperaPiPToggleSingleInstanceMutex")
	if err != nil {
		return 0, err
	}
	ret, _, callErr := procCreateMutexW.Call(0, 1, uintptr(unsafe.Pointer(mutexName)))
	if ret == 0 {
		return 0, fmt.Errorf("CreateMutexW failed: %v", callErr)
	}
	if callErr == syscall.Errno(ERROR_ALREADY_EXISTS) {
		res, _, _ := procWaitForSingleObject.Call(ret, 0)
		if res == uintptr(WAIT_TIMEOUT) {
			procCloseHandle.Call(ret)
			return 0, fmt.Errorf("another instance is already running")
		}
	}
	return ret, nil
}

func main() {
	// Setup dual logging to absolute path next to executable
	execPath, errPath := os.Executable()
	var logPath string
	if errPath == nil {
		logPath = filepath.Join(filepath.Dir(execPath), "pip-toggle.log")
	} else {
		logPath = "pip-toggle.log"
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
	} else {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	log.Println("==================================================")
	log.Println(" PiPs Click-Through Utility Starting...")
	log.Println("==================================================")
	if configDir, err := os.UserConfigDir(); err != nil {
		log.Printf("[main] Settings directory unavailable: %v", err)
	} else {
		languageSettingsPath = filepath.Join(configDir, "PiPs", "settings.json")
		var err error
		uiLanguage, err = i18n.LoadLanguage(languageSettingsPath)
		if err != nil {
			log.Printf("[main] Could not load language; using Russian: %v", err)
		}
	}

	mutexHandle, err := checkSingleInstance()
	if err != nil {
		log.Println("[main] Application is already running in background!")
		notify.ShowNotification("PiPs", tr(i18n.AlreadyRunning))
		return
	}
	defer procCloseHandle.Call(mutexHandle)

	systray.Run(onReady, onExit)
}

func updateStateUI(isClickThrough bool) {
	currentStatus = isClickThrough
	if isClickThrough {
		systray.SetIcon(icon.GetIconOn())
		systray.SetTooltip(tr(i18n.TooltipOn))
		if mStatus != nil {
			mStatus.SetTitle(tr(i18n.StatusOn))
		}
	} else {
		systray.SetIcon(icon.GetIconOff())
		systray.SetTooltip(tr(i18n.TooltipOff))
		if mStatus != nil {
			mStatus.SetTitle(tr(i18n.StatusOff))
		}
	}
}

func handleToggle() {
	log.Println("[main] Hotkey/Menu: Toggle Click-Through requested")

	hwnd, title, err := winapi.FindPiPWindow()
	if err != nil {
		log.Printf("[main] Target isolation check: %v", err)
		updateStateUI(false)
		notify.ShowNotification("PiPs", tr(i18n.WindowNotFound))
		return
	}

	log.Printf("[main] Target isolation success: HWND=0x%X Title=%q", hwnd, title)

	newState, err := winapi.ToggleClickThrough(hwnd)
	if err != nil {
		log.Printf("[main] Error toggling click-through: %v", err)
		notify.ShowNotification("PiPs", tr(i18n.ToggleError))
		return
	}

	updateStateUI(newState)

	if newState {
		log.Println("[main] Click-through ENABLED on PiP window")
		notify.ShowNotification("PiPs", tr(i18n.ClicksOn))
	} else {
		log.Println("[main] Click-through DISABLED on PiP window")
		notify.ShowNotification("PiPs", tr(i18n.ClicksOff))
	}
}

func handleEmergencyReset() {
	resetPiPClicks(winapi.FindPiPWindow, winapi.SetClickThrough, winapi.IsClickThrough, updateStateUI, notify.ShowNotification)
}

func resetPiPClicks(
	findWindow func() (uintptr, string, error),
	setClickThrough func(uintptr, bool) (bool, error),
	isClickThrough func(uintptr) bool,
	updateState func(bool),
	showNotification func(string, string),
) {
	log.Println("[main] Emergency Reset requested")

	hwnd, title, err := findWindow()
	if err != nil {
		log.Printf("[main] Emergency Reset: PiP window not found (%v)", err)
		updateState(false)
		showNotification(tr(i18n.ResetTitle), tr(i18n.WindowNotFound))
		return
	}

	log.Printf("[main] Emergency Reset target: HWND=0x%X Title=%q", hwnd, title)
	_, err = setClickThrough(hwnd, false)
	if err != nil {
		log.Printf("[main] Emergency Reset failed: %v", err)
		updateState(isClickThrough(hwnd))
		showNotification(tr(i18n.ResetTitle), tr(i18n.ResetError))
		return
	}

	updateState(false)
	showNotification(tr(i18n.ResetTitle), tr(i18n.ResetSuccess))
}

func onReady() {
	updateStateUI(false)

	// Start permanent PiP background monitor for auto-rounding
	winapi.StartPermanentPiPMonitor()

	mStatus = systray.AddMenuItem(tr(i18n.StatusOff), tr(i18n.StatusTip))
	mStatus.Disable()

	systray.AddSeparator()

	mToggle := systray.AddMenuItem(tr(i18n.Toggle), tr(i18n.ToggleTip))
	mPlayPause := systray.AddMenuItem(tr(i18n.PlayPause), tr(i18n.PlayPauseTip))
	mRoundToggle := systray.AddMenuItem(tr(i18n.RoundingOn), tr(i18n.RoundingTip))

	mAutostart := systray.AddMenuItem(tr(i18n.AutostartOn), tr(i18n.AutostartTip))
	if !autostart.IsEnabled() {
		mAutostart.SetTitle(tr(i18n.AutostartOff))
	}

	mReset := systray.AddMenuItem(tr(i18n.Reset), tr(i18n.ResetTip))
	mLanguage := systray.AddMenuItem(tr(i18n.LanguageMenu), tr(i18n.LanguageTip))
	mRussian := mLanguage.AddSubMenuItemCheckbox("Русский", "Русский", uiLanguage == i18n.Russian)
	mEnglish := mLanguage.AddSubMenuItemCheckbox("English", "English", uiLanguage == i18n.English)
	mUkrainian := mLanguage.AddSubMenuItemCheckbox("Українська", "Українська", uiLanguage == i18n.Ukrainian)

	systray.AddSeparator()

	mExit := systray.AddMenuItem(tr(i18n.Exit), tr(i18n.ExitTip))
	refreshLanguage := func() {
		updateStateUI(currentStatus)
		mStatus.SetTooltip(tr(i18n.StatusTip))
		setMenuText(mToggle, i18n.Toggle, i18n.ToggleTip)
		setMenuText(mPlayPause, i18n.PlayPause, i18n.PlayPauseTip)
		setMenuText(mRoundToggle, stateKey(winapi.IsGlobalRoundingEnabled(), i18n.RoundingOn, i18n.RoundingOff), i18n.RoundingTip)
		setMenuText(mAutostart, stateKey(autostart.IsEnabled(), i18n.AutostartOn, i18n.AutostartOff), i18n.AutostartTip)
		setMenuText(mReset, i18n.Reset, i18n.ResetTip)
		setMenuText(mLanguage, i18n.LanguageMenu, i18n.LanguageTip)
		setMenuText(mExit, i18n.Exit, i18n.ExitTip)
		for language, item := range map[i18n.Language]*systray.MenuItem{i18n.Russian: mRussian, i18n.English: mEnglish, i18n.Ukrainian: mUkrainian} {
			if language == uiLanguage {
				item.Check()
			} else {
				item.Uncheck()
			}
		}
	}
	selectLanguage := func(language i18n.Language) {
		if err := changeUILanguage(language, func(value i18n.Language) error {
			return i18n.SaveLanguage(languageSettingsPath, value)
		}); err != nil {
			log.Printf("[main] Could not save language: %v", err)
			notify.ShowNotification("PiPs", tr(i18n.LanguageSaveError))
			return
		}
		refreshLanguage()
	}

	// Initialize global hotkey listener
	hkListener := hotkey.NewListener()
	if err := hkListener.Start(); err != nil {
		log.Printf("[main] Critical error starting hotkeys: %v", err)
		notify.ShowNotification("PiPs", tr(i18n.HotkeyError))
	}
	defer hkListener.Stop()

	// OS Shutdown Signal Handler
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	chToggle := mToggle.ClickedCh
	chPlayPause := mPlayPause.ClickedCh
	chRoundToggle := mRoundToggle.ClickedCh
	chAutostart := mAutostart.ClickedCh
	chReset := mReset.ClickedCh
	chExit := mExit.ClickedCh
	chRussian := mRussian.ClickedCh
	chEnglish := mEnglish.ClickedCh
	chUkrainian := mUkrainian.ClickedCh

	for {
		select {
		case _, ok := <-chRussian:
			if !ok {
				chRussian = nil
				continue
			}
			selectLanguage(i18n.Russian)
		case _, ok := <-chEnglish:
			if !ok {
				chEnglish = nil
				continue
			}
			selectLanguage(i18n.English)
		case _, ok := <-chUkrainian:
			if !ok {
				chUkrainian = nil
				continue
			}
			selectLanguage(i18n.Ukrainian)
		case hkID, ok := <-hkListener.Events:
			if !ok {
				hkListener.Events = nil
				continue
			}
			if hkID == hotkey.ID_TOGGLE {
				handleToggle()
			} else if hkID == hotkey.ID_RESET {
				handleEmergencyReset()
			} else if hkID == hotkey.ID_PLAY_PAUSE {
				log.Println("[main] Hotkey: Play/Pause requested")
				winapi.SendPlayPause()
			}

		case _, ok := <-chToggle:
			if !ok {
				chToggle = nil
				continue
			}
			handleToggle()

		case _, ok := <-chPlayPause:
			if !ok {
				chPlayPause = nil
				continue
			}
			log.Println("[main] Menu: Play/Pause requested")
			winapi.SendPlayPause()

		case _, ok := <-chRoundToggle:
			if !ok {
				chRoundToggle = nil
				continue
			}
			newState := !winapi.IsGlobalRoundingEnabled()
			winapi.SetGlobalRoundingEnabled(newState)
			if newState {
				mRoundToggle.SetTitle(tr(i18n.RoundingOn))
				log.Println("[main] Menu: Global PiP Rounding ENABLED")
			} else {
				mRoundToggle.SetTitle(tr(i18n.RoundingOff))
				log.Println("[main] Menu: Global PiP Rounding DISABLED")
			}

		case _, ok := <-chAutostart:
			if !ok {
				chAutostart = nil
				continue
			}
			newState := !autostart.IsEnabled()
			if err := autostart.SetEnabled(newState); err != nil {
				log.Printf("[main] Error toggling autostart: %v", err)
				notify.ShowNotification("PiPs", tr(i18n.AutostartError))
			} else {
				if newState {
					mAutostart.SetTitle(tr(i18n.AutostartOn))
					notify.ShowNotification("PiPs", tr(i18n.AutostartEnabled))
				} else {
					mAutostart.SetTitle(tr(i18n.AutostartOff))
					notify.ShowNotification("PiPs", tr(i18n.AutostartDisabled))
				}
			}

		case _, ok := <-chReset:
			if !ok {
				chReset = nil
				continue
			}
			handleEmergencyReset()

		case _, ok := <-chExit:
			if !ok {
				chExit = nil
				continue
			}
			log.Println("[main] Exit menu item clicked")
			cleanupAndExit()
			return

		case sig := <-sigCh:
			log.Printf("[main] Received OS signal: %v", sig)
			cleanupAndExit()
			return
		}
	}
}

func cleanupAndExit() {
	log.Println("[main] Performing emergency cleanup on exit...")
	winapi.StopPermanentPiPMonitor()
	if err := winapi.RestoreClickThroughWindows(); err != nil {
		log.Printf("[main] Failed to restore modified PiP windows: %v", err)
	}
	hwnd, title, err := winapi.FindPiPWindow()
	if err == nil && hwnd != 0 {
		log.Printf("[main] Restoring normal mouse behavior for HWND=0x%X (%q)...", hwnd, title)
		if _, err := winapi.SetClickThrough(hwnd, false); err != nil {
			log.Printf("[main] Failed to restore clicks on exit: %v", err)
		}
	} else {
		log.Println("[main] No active PiP window found during cleanup.")
	}
	log.Println("[main] Exiting application cleanly.")
	systray.Quit()
}

func onExit() {
	log.Println("[main] Systray finished. Goodbye.")
}
