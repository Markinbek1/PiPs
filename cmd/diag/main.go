package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"pip-toggle/pkg/winapi"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procEnumWindows              = user32.NewProc("EnumWindows")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetClassNameW            = user32.NewProc("GetClassNameW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procGetWindowRect            = user32.NewProc("GetWindowRect")

	procGetWindowLongPtrW *syscall.LazyProc

	procOpenProcess               = kernel32.NewProc("OpenProcess")
	procCloseHandle               = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
)

func init() {
	proc := user32.NewProc("GetWindowLongPtrW")
	if err := proc.Find(); err != nil {
		proc = user32.NewProc("GetWindowLongW")
	}
	procGetWindowLongPtrW = proc
}

const (
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	GWL_EXSTYLE                       = -20
)

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type WindowInfo struct {
	HWND        uintptr
	PID         uint32
	ProcessName string
	ClassName   string
	Title       string
	ExStyle     uintptr
	Rect        RECT
	Width       int32
	Height      int32
	Visible     bool
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

func getWindowRect(hwnd uintptr) (RECT, int32, int32) {
	var rect RECT
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top
	return rect, width, height
}

func isVisible(hwnd uintptr) bool {
	ret, _, _ := procIsWindowVisible.Call(hwnd)
	return ret != 0
}

func main() {
	fmt.Println("==================================================")
	fmt.Println(" PiPs Windows Diagnostic Utility")
	fmt.Println(" Searching for visible Opera, Chrome and Edge windows...")
	fmt.Println("==================================================")

	var browserWindows []WindowInfo

	cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		if !isVisible(hwnd) {
			return 1 // continue enumeration
		}

		var pid uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if pid == 0 {
			return 1
		}

		procName := getProcessName(pid)
		if !winapi.IsSupportedBrowserProcess(procName) {
			return 1
		}

		className := getClassName(hwnd)
		title := getWindowText(hwnd)
		exStyle := getExStyle(hwnd)
		rect, width, height := getWindowRect(hwnd)

		info := WindowInfo{
			HWND:        hwnd,
			PID:         pid,
			ProcessName: procName,
			ClassName:   className,
			Title:       title,
			ExStyle:     exStyle,
			Rect:        rect,
			Width:       width,
			Height:      height,
			Visible:     true,
		}

		browserWindows = append(browserWindows, info)
		return 1
	})

	procEnumWindows.Call(cb, 0)

	if len(browserWindows) == 0 {
		fmt.Println("\nNo visible windows belonging to Opera, Chrome or Edge were found.")
		fmt.Println("Open a supported browser and run the diagnostic utility again.")
		return
	}

	fmt.Printf("\nFound %d visible browser window(s):\n\n", len(browserWindows))

	for i, win := range browserWindows {
		fmt.Printf("[%d] HWND: 0x%X (%d)\n", i+1, win.HWND, win.HWND)
		fmt.Printf("    PID:         %d\n", win.PID)
		fmt.Printf("    Class Name:  %q\n", win.ClassName)
		fmt.Printf("    Title:       %q\n", win.Title)
		fmt.Printf("    ExStyle:     0x%08X\n", win.ExStyle)
		fmt.Printf("    Bounds:      Left=%d, Top=%d, Right=%d, Bottom=%d (Width=%d, Height=%d)\n",
			win.Rect.Left, win.Rect.Top, win.Rect.Right, win.Rect.Bottom, win.Width, win.Height)
		fmt.Println(strings.Repeat("-", 60))
	}
}
