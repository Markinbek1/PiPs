package winapi

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procEnumWindows              = user32.NewProc("EnumWindows")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procIsWindow                 = user32.NewProc("IsWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetClassNameW            = user32.NewProc("GetClassNameW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procGetWindowRgn             = user32.NewProc("GetWindowRgn")

	procGetWindowLongPtrW   *syscall.LazyProc
	procSetWindowLongPtrW   *syscall.LazyProc
	procSendMessageTimeoutW = user32.NewProc("SendMessageTimeoutW")
	procKeybdEvent          = user32.NewProc("keybd_event")
	procSetWindowRgn        = user32.NewProc("SetWindowRgn")
	procSetWinEventHook     = user32.NewProc("SetWinEventHook")
	procUnhookWinEvent      = user32.NewProc("UnhookWinEvent")

	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	procCreateRoundRectRgn = gdi32.NewProc("CreateRoundRectRgn")
	procCreateRectRgn      = gdi32.NewProc("CreateRectRgn")
	procDeleteObject       = gdi32.NewProc("DeleteObject")

	dwmapi                    = syscall.NewLazyDLL("dwmapi.dll")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
	procDwmGetWindowAttribute = dwmapi.NewProc("DwmGetWindowAttribute")

	procOpenProcess               = kernel32.NewProc("OpenProcess")
	procSetLastError              = kernel32.NewProc("SetLastError")
	procCloseHandle               = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
)

const (
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	GWL_EXSTYLE                       = -20

	WS_EX_TOPMOST     = 0x00000008
	WS_EX_TRANSPARENT = 0x00000020
	WS_EX_LAYERED     = 0x00080000

	VK_MEDIA_PLAY_PAUSE         = 0xB3
	WM_APPCOMMAND               = 0x0319
	APPCOMMAND_MEDIA_PLAY_PAUSE = 14
	KEYEVENTF_KEYUP             = 2

	DWMWA_WINDOW_CORNER_PREFERENCE = 33
	DWMWCP_ROUND                   = 2
)

var (
	enumCbMutex sync.Mutex
	foundHWND   uintptr
	foundTitle  string
	enumCbPtr   uintptr
)

func enumWindowsCallback(hwnd uintptr, lparam uintptr) uintptr {
	if !isVisible(hwnd) {
		return 1 // continue searching
	}

	// Cheap checks run before opening a process handle.
	className := getClassName(hwnd)
	exStyle := getExStyle(hwnd)
	if !hasPiPWindowStyle(className, exStyle) {
		return 1
	}

	// A regular tab mentioning PiP must not be mistaken for the native overlay.
	title := getWindowText(hwnd)
	if !IsPiPTitle(title) {
		return 1
	}

	// Electron/WebView apps can share the window class, so keep a process allowlist.
	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return 1
	}

	procName := getProcessName(pid)
	if !IsSupportedBrowserProcess(procName) {
		return 1
	}

	// Found matching PiP window!
	foundHWND = hwnd
	foundTitle = title
	return 0 // stop enumeration
}

func init() {
	procGet := user32.NewProc("GetWindowLongPtrW")
	if err := procGet.Find(); err != nil {
		procGet = user32.NewProc("GetWindowLongW")
	}
	procGetWindowLongPtrW = procGet

	procSet := user32.NewProc("SetWindowLongPtrW")
	if err := procSet.Find(); err != nil {
		procSet = user32.NewProc("SetWindowLongW")
	}
	procSetWindowLongPtrW = procSet

	enumCbPtr = syscall.NewCallback(enumWindowsCallback)
}

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

func getProcessName(pid uint32) string {
	handle, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_LIMITED_INFORMATION),
		uintptr(0),
		uintptr(pid),
	)
	if handle == 0 {
		return ""
	}
	defer procCloseHandle.Call(handle)

	var buf [1024]uint16
	size := uint32(len(buf))
	ret, _, _ := procQueryFullProcessImageName.Call(
		handle,
		uintptr(0),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return ""
	}

	fullPath := syscall.UTF16ToString(buf[:size])
	return filepath.Base(fullPath)
}

func getClassName(hwnd uintptr) string {
	var buf [256]uint16
	ret, _, _ := procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:ret])
}

func getWindowText(hwnd uintptr) string {
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	ret, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:ret])
}

func getExStyle(hwnd uintptr) uintptr {
	var gwl int32 = GWL_EXSTYLE
	ret, _, _ := procGetWindowLongPtrW.Call(hwnd, uintptr(gwl))
	return ret
}

func setExStyle(hwnd uintptr, newStyle uintptr) bool {
	// A zero previous style is also a successful return; check GetLastError.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	procSetLastError.Call(0)
	var gwl int32 = GWL_EXSTYLE
	ret, _, callErr := procSetWindowLongPtrW.Call(hwnd, uintptr(gwl), newStyle)
	return ret != 0 || callErr == syscall.Errno(0)
}

func isVisible(hwnd uintptr) bool {
	ret, _, _ := procIsWindowVisible.Call(hwnd)
	return ret != 0
}

// FindPiPWindow finds the first visible native PiP overlay in window order.
// Class, topmost style, full title and browser process are all required.
func FindPiPWindow() (uintptr, string, error) {
	enumCbMutex.Lock()
	defer enumCbMutex.Unlock()

	foundHWND = 0
	foundTitle = ""

	procEnumWindows.Call(enumCbPtr, 0)

	if foundHWND == 0 {
		return 0, "", fmt.Errorf("PiP window not found in Opera, Chrome or Edge")
	}

	return foundHWND, foundTitle, nil
}

// IsClickThrough returns true if WS_EX_TRANSPARENT flag is set on the window.
func IsClickThrough(hwnd uintptr) bool {
	exStyle := getExStyle(hwnd)
	return exStyle&WS_EX_TRANSPARENT != 0
}

func getWindowRect(hwnd uintptr) (RECT, int32, int32) {
	var rect RECT
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top
	return rect, width, height
}

const (
	EVENT_OBJECT_LOCATIONCHANGE = 0x800B
	WINEVENT_OUTOFCONTEXT       = 0x0000
)

var (
	roundingMu        sync.Mutex
	rounding          = roundingState{enabled: true}
	autoResizeDoneCh  chan struct{}
	autoResizeStopped chan struct{}
)

// SetGlobalRoundingEnabled enables or disables automatic rounding globally
func SetGlobalRoundingEnabled(enable bool) {
	roundingMu.Lock()
	defer roundingMu.Unlock()
	rounding.setEnabled(enable, ResetRoundedCorners)
}

func IsGlobalRoundingEnabled() bool {
	roundingMu.Lock()
	defer roundingMu.Unlock()
	return rounding.enabled
}

func absInt32(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

// StartPermanentPiPMonitor keeps rounding applied to the current browser PiP.
func StartPermanentPiPMonitor() {
	roundingMu.Lock()
	defer roundingMu.Unlock()
	if autoResizeDoneCh != nil {
		return
	}

	autoResizeDoneCh = make(chan struct{})
	autoResizeStopped = make(chan struct{})
	done := autoResizeDoneCh
	stopped := autoResizeStopped

	go func() {
		defer close(stopped)
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[winapi] Panic recovered in PiPMonitor: %v", r)
			}
		}()
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				updateRoundedWindow()
			}
		}
	}()

	log.Println("[winapi] Started permanent browser PiP background monitor")
}

func updateRoundedWindow() {
	roundingMu.Lock()
	defer roundingMu.Unlock()
	if !rounding.enabled {
		return
	}
	hwnd, _, err := FindPiPWindow()
	if err != nil || hwnd == 0 {
		rounding.update(0, 0, 0, ApplyRoundedCorners, ResetRoundedCorners)
		return
	}
	_, width, height := getWindowRect(hwnd)
	rounding.update(hwnd, width, height, ApplyRoundedCorners, ResetRoundedCorners)
}

// StopPermanentPiPMonitor waits for the worker before restoring the last window.
func StopPermanentPiPMonitor() {
	roundingMu.Lock()
	stopped := autoResizeStopped
	if autoResizeDoneCh != nil {
		close(autoResizeDoneCh)
		autoResizeDoneCh = nil
	}
	roundingMu.Unlock()
	if stopped != nil {
		<-stopped
	}
	roundingMu.Lock()
	defer roundingMu.Unlock()
	rounding.reset(ResetRoundedCorners)
}

type cornerState struct {
	dwmPreference  uint32
	usesDWM        bool
	originalRegion uintptr
}

var (
	cornerMu       sync.Mutex
	roundedWindows = make(map[uintptr]cornerState)
)

// ApplyRoundedCorners clips/rounds the HWND corners with a smooth radius (e.g., 24px)
func ApplyRoundedCorners(hwnd uintptr, radius int) bool {
	if hwnd == 0 {
		return false
	}
	cornerMu.Lock()
	defer cornerMu.Unlock()

	// Try Windows 11 DWM rounded corners first if available
	if err := procDwmSetWindowAttribute.Find(); err == nil {
		state, saved := roundedWindows[hwnd]
		if !saved {
			pref := uint32(0)
			ret, _, _ := procDwmGetWindowAttribute.Call(hwnd, uintptr(DWMWA_WINDOW_CORNER_PREFERENCE), uintptr(unsafe.Pointer(&pref)), unsafe.Sizeof(pref))
			if ret == 0 {
				state = cornerState{dwmPreference: pref, usesDWM: true}
			}
		}
		if state.usesDWM {
			pref := uint32(DWMWCP_ROUND)
			ret, _, _ := procDwmSetWindowAttribute.Call(hwnd, uintptr(DWMWA_WINDOW_CORNER_PREFERENCE), uintptr(unsafe.Pointer(&pref)), unsafe.Sizeof(pref))
			if ret == 0 {
				roundedWindows[hwnd] = state
				log.Printf("[winapi] DWM rounded corners applied to HWND 0x%X", hwnd)
				return true
			}
			// Keep the original preference until a later retry can restore it.
			return false
		}
	}

	// Windows 10 GDI SetWindowRgn rounding
	_, width, height := getWindowRect(hwnd)
	if width > 0 && height > 0 {
		state, saved := roundedWindows[hwnd]
		if !saved {
			original, _, _ := procCreateRectRgn.Call(0, 0, 0, 0)
			if original == 0 {
				return false
			}
			ret, _, _ := procGetWindowRgn.Call(hwnd, original)
			if ret == 0 {
				procDeleteObject.Call(original)
			} else {
				state.originalRegion = original
			}
		}
		rgn, _, _ := procCreateRoundRectRgn.Call(
			0, 0,
			uintptr(width+1), uintptr(height+1),
			uintptr(radius), uintptr(radius),
		)
		if rgn != 0 {
			// Pass bRedraw = 0 so Windows doesn't spam Chromium with disruptive layout redraws
			ret, _, _ := procSetWindowRgn.Call(hwnd, rgn, 0)
			if ret != 0 {
				roundedWindows[hwnd] = state
				log.Printf("[winapi] Windows 10 GDI rounded corners set for HWND 0x%X", hwnd)
				return true
			}
			// Ownership only passes to Windows when SetWindowRgn succeeds.
			procDeleteObject.Call(rgn)
		}
		if !saved && state.originalRegion != 0 {
			procDeleteObject.Call(state.originalRegion)
		}
	}
	return false
}

// ResetRoundedCorners restores the DWM preference or removes our GDI clipping.
func ResetRoundedCorners(hwnd uintptr) {
	cornerMu.Lock()
	defer cornerMu.Unlock()
	state, saved := roundedWindows[hwnd]
	if !saved {
		return
	}
	valid, _, _ := procIsWindow.Call(hwnd)
	if valid == 0 {
		if state.originalRegion != 0 {
			procDeleteObject.Call(state.originalRegion)
		}
		delete(roundedWindows, hwnd)
		return
	}
	if state.usesDWM {
		pref := state.dwmPreference
		ret, _, _ := procDwmSetWindowAttribute.Call(hwnd, uintptr(DWMWA_WINDOW_CORNER_PREFERENCE), uintptr(unsafe.Pointer(&pref)), unsafe.Sizeof(pref))
		if ret != 0 {
			log.Printf("[winapi] Failed to restore DWM corners for HWND 0x%X", hwnd)
			return
		}
	} else {
		ret, _, _ := procSetWindowRgn.Call(hwnd, state.originalRegion, 1)
		if ret == 0 {
			log.Printf("[winapi] Failed to restore GDI corners for HWND 0x%X", hwnd)
			return
		}
	}
	delete(roundedWindows, hwnd)
	log.Printf("[winapi] Reset rounded corners for HWND 0x%X", hwnd)
}

var (
	clickThroughMu sync.Mutex
	addedLayered   = make(map[uintptr]bool)
)

// SetClickThrough sets or removes mouse click-through capability for the target HWND.
func SetClickThrough(hwnd uintptr, enable bool) (bool, error) {
	clickThroughMu.Lock()
	defer clickThroughMu.Unlock()
	valid, _, _ := procIsWindow.Call(hwnd)
	if valid == 0 {
		delete(addedLayered, hwnd)
		return false, fmt.Errorf("PiP window no longer exists")
	}
	currentStyle := getExStyle(hwnd)
	var newStyle uintptr

	if enable {
		newStyle = currentStyle | WS_EX_TRANSPARENT | WS_EX_LAYERED
	} else {
		newStyle = currentStyle &^ WS_EX_TRANSPARENT
		if addedLayered[hwnd] {
			newStyle &^= WS_EX_LAYERED
		}
	}

	if currentStyle != newStyle {
		success := setExStyle(hwnd, newStyle)
		if !success {
			log.Printf("[winapi] Failed SetWindowLongPtrW for HWND 0x%X", hwnd)
			return false, fmt.Errorf("failed to change window style")
		}
		log.Printf("[winapi] HWND 0x%X style changed: 0x%08X -> 0x%08X (ClickThrough=%v)", hwnd, currentStyle, newStyle, enable)
	}
	if enable {
		if _, tracked := addedLayered[hwnd]; !tracked {
			addedLayered[hwnd] = currentStyle&WS_EX_LAYERED == 0
		}
	} else if !enable {
		delete(addedLayered, hwnd)
	}

	return enable, nil
}

// RestoreClickThroughWindows also restores overlays from another browser that
// were toggled earlier, even if a different PiP window is now first in the list.
func RestoreClickThroughWindows() error {
	clickThroughMu.Lock()
	handles := make([]uintptr, 0, len(addedLayered))
	for hwnd := range addedLayered {
		handles = append(handles, hwnd)
	}
	clickThroughMu.Unlock()
	var failures []error
	for _, hwnd := range handles {
		valid, _, _ := procIsWindow.Call(hwnd)
		if valid == 0 {
			clickThroughMu.Lock()
			delete(addedLayered, hwnd)
			clickThroughMu.Unlock()
			continue
		}
		if _, err := SetClickThrough(hwnd, false); err != nil {
			failures = append(failures, fmt.Errorf("HWND 0x%X: %w", hwnd, err))
		}
	}
	return errors.Join(failures...)
}

// ToggleClickThrough toggles click-through state on the target HWND.
func ToggleClickThrough(hwnd uintptr) (bool, error) {
	currentState := IsClickThrough(hwnd)
	newState := !currentState
	return SetClickThrough(hwnd, newState)
}

// SendPlayPause sends a play/pause command to browser PiP or a media-key fallback.
func SendPlayPause() {
	sendPlayPause(FindPiPWindow, func(hwnd uintptr) bool {
		var result uintptr
		ret, _, _ := procSendMessageTimeoutW.Call(hwnd, WM_APPCOMMAND, hwnd, uintptr(APPCOMMAND_MEDIA_PLAY_PAUSE<<16), 0x0003, 1000, uintptr(unsafe.Pointer(&result)))
		return ret != 0 && result != 0
	}, func() {
		procKeybdEvent.Call(uintptr(VK_MEDIA_PLAY_PAUSE), 0, 0, 0)
		procKeybdEvent.Call(uintptr(VK_MEDIA_PLAY_PAUSE), 0, uintptr(KEYEVENTF_KEYUP), 0)
	})
}

func sendPlayPause(findWindow func() (uintptr, string, error), sendToWindow func(uintptr) bool, sendMediaKey func()) {
	hwnd, _, err := findWindow()
	if err == nil && hwnd != 0 && sendToWindow(hwnd) {
		return
	}
	sendMediaKey()
}
