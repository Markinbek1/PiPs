package hotkey

import (
	"fmt"
	"log"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterHotKey    = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey  = user32.NewProc("UnregisterHotKey")
	procGetMessageW       = user32.NewProc("GetMessageW")
	procPostThreadMessage = user32.NewProc("PostThreadMessageW")
	procGetCurrentThread  = kernel32.NewProc("GetCurrentThreadId")
)

const (
	WM_HOTKEY = 0x0312
	WM_QUIT   = 0x0012

	MOD_ALT      = 0x0001
	MOD_CONTROL  = 0x0002
	MOD_SHIFT    = 0x0004
	MOD_NOREPEAT = 0x4000

	VK_P     = 0x50 // Virtual key code for 'P'
	VK_SPACE = 0x20 // Virtual key code for Space

	ID_TOGGLE     = 1
	ID_RESET      = 2
	ID_PLAY_PAUSE = 3
)

type MSG struct {
	HWND    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type Listener struct {
	Events   chan int
	done     chan struct{}
	threadID uint32
}

func NewListener() *Listener {
	return &Listener{
		Events: make(chan int, 10),
		done:   make(chan struct{}),
	}
}

func (l *Listener) Start() error {
	errCh := make(chan error, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		tid, _, _ := procGetCurrentThread.Call()
		l.threadID = uint32(tid)

		// 1. Register Ctrl + Alt + P (Toggle)
		ret1, _, err1 := procRegisterHotKey.Call(
			0,
			uintptr(ID_TOGGLE),
			uintptr(MOD_CONTROL|MOD_ALT|MOD_NOREPEAT),
			uintptr(VK_P),
		)
		if ret1 == 0 {
			errCh <- fmt.Errorf("failed to register Ctrl+Alt+P hotkey: %v", err1)
			return
		}

		// 2. Register Ctrl + Alt + Shift + P (Emergency Reset)
		ret2, _, err2 := procRegisterHotKey.Call(
			0,
			uintptr(ID_RESET),
			uintptr(MOD_CONTROL|MOD_ALT|MOD_SHIFT|MOD_NOREPEAT),
			uintptr(VK_P),
		)
		if ret2 == 0 {
			procUnregisterHotKey.Call(0, uintptr(ID_TOGGLE))
			errCh <- fmt.Errorf("failed to register Ctrl+Alt+Shift+P hotkey: %v", err2)
			return
		}

		// 3. Register Ctrl + Alt + Space (Play/Pause Media)
		ret3, _, err3 := procRegisterHotKey.Call(
			0,
			uintptr(ID_PLAY_PAUSE),
			uintptr(MOD_CONTROL|MOD_ALT|MOD_NOREPEAT),
			uintptr(VK_SPACE),
		)
		if ret3 == 0 {
			log.Printf("[hotkey] Warning: failed to register Ctrl+Alt+Space hotkey (%v), play/pause hotkey will be unavailable", err3)
		}

		log.Println("[hotkey] Registered hotkeys successfully: Ctrl+Alt+P (Toggle), Ctrl+Alt+Shift+P (Reset), Ctrl+Alt+Space (Play/Pause)")
		errCh <- nil

		var msg MSG
		for {
			res, _, _ := procGetMessageW.Call(
				uintptr(unsafe.Pointer(&msg)),
				0,
				0,
				0,
			)
			if int32(res) <= 0 {
				break
			}

			if msg.Message == WM_HOTKEY {
				hotkeyID := int(msg.WParam)
				select {
				case l.Events <- hotkeyID:
				default:
				}
			}
		}

		procUnregisterHotKey.Call(0, uintptr(ID_TOGGLE))
		procUnregisterHotKey.Call(0, uintptr(ID_RESET))
		procUnregisterHotKey.Call(0, uintptr(ID_PLAY_PAUSE))
		log.Println("[hotkey] Hotkey thread exited cleanly.")
	}()

	return <-errCh
}

func (l *Listener) Stop() {
	if l.threadID != 0 {
		procPostThreadMessage.Call(uintptr(l.threadID), uintptr(WM_QUIT), 0, 0)
	}
}
