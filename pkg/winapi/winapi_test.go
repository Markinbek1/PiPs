package winapi

import (
	"errors"
	"runtime"
	"syscall"
	"testing"
	"unsafe"
)

func TestPlayPauseUsesOneRoute(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		hwnd                  uintptr
		err                   error
		handled               bool
		wantWindow, wantMedia int
	}{
		{"handled by PiP", 42, nil, true, 1, 0},
		{"unhandled by PiP", 42, nil, false, 1, 1},
		{"no PiP window", 0, errors.New("not found"), false, 0, 1},
		{"empty handle", 0, nil, false, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			windowCalls, mediaCalls := 0, 0
			sendPlayPause(func() (uintptr, string, error) { return tc.hwnd, "PiP", tc.err }, func(hwnd uintptr) bool {
				if hwnd != tc.hwnd {
					t.Fatalf("unexpected target: %d", hwnd)
				}
				windowCalls++
				return tc.handled
			}, func() { mediaCalls++ })
			if windowCalls != tc.wantWindow || mediaCalls != tc.wantMedia {
				t.Fatalf("window=%d media=%d, want window=%d media=%d", windowCalls, mediaCalls, tc.wantWindow, tc.wantMedia)
			}
		})
	}
}

// Tests own a hidden window; they never touch Opera or send real media keys.
func hiddenTestWindow(t *testing.T) uintptr {
	t.Helper()
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)
	class, _ := syscall.UTF16PtrFromString("STATIC")
	title, _ := syscall.UTF16PtrFromString("PiPs regression test")
	hwnd, _, err := user32.NewProc("CreateWindowExW").Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(title)), 0x80000000, 0, 0, 640, 360, 0, 0, 0, 0)
	if hwnd == 0 {
		t.Fatalf("create hidden window: %v", err)
	}
	t.Cleanup(func() { user32.NewProc("DestroyWindow").Call(hwnd) })
	return hwnd
}

func TestClickThroughRestoresAddedStyles(t *testing.T) {
	hwnd := hiddenTestWindow(t)
	for _, original := range []uintptr{0, WS_EX_LAYERED, 0x00000080} {
		if !setExStyle(hwnd, original) {
			t.Fatalf("set initial style 0x%X", original)
		}
		if _, err := SetClickThrough(hwnd, true); err != nil {
			t.Fatal(err)
		}
		if got := getExStyle(hwnd); got != original|WS_EX_TRANSPARENT|WS_EX_LAYERED {
			t.Fatalf("enabled style=0x%X", got)
		}
		// Repeated enabling must not lose knowledge of a style we added.
		if _, err := SetClickThrough(hwnd, true); err != nil {
			t.Fatal(err)
		}
		if _, err := SetClickThrough(hwnd, false); err != nil {
			t.Fatal(err)
		}
		if got := getExStyle(hwnd); got != original {
			t.Fatalf("restored style=0x%X, want 0x%X", got, original)
		}
	}
}

func TestClickThroughRejectsInvalidWindow(t *testing.T) {
	for _, enable := range []bool{true, false} {
		if _, err := SetClickThrough(0, enable); err == nil {
			t.Fatal("invalid HWND must return an error, including during emergency reset")
		}
	}
}

func TestExitRestoresEveryModifiedOverlay(t *testing.T) {
	first := hiddenTestWindow(t)
	second := hiddenTestWindow(t)
	untouched := hiddenTestWindow(t)
	if !setExStyle(second, WS_EX_LAYERED) || !setExStyle(untouched, WS_EX_LAYERED|WS_EX_TRANSPARENT) {
		t.Fatal("failed to set original test styles")
	}
	if _, err := SetClickThrough(first, true); err != nil {
		t.Fatal(err)
	}
	if _, err := SetClickThrough(second, true); err != nil {
		t.Fatal(err)
	}
	if err := RestoreClickThroughWindows(); err != nil {
		t.Fatal(err)
	}
	if getExStyle(first) != 0 || getExStyle(second) != WS_EX_LAYERED {
		t.Fatal("all modified windows must be restored, including already layered windows")
	}
	if getExStyle(untouched) != WS_EX_LAYERED|WS_EX_TRANSPARENT {
		t.Fatal("cleanup must leave unrelated windows unchanged")
	}
}

func TestRoundedCornersRestoreOriginalPreference(t *testing.T) {
	hwnd := hiddenTestWindow(t)
	var original uint32
	ret, _, _ := procDwmGetWindowAttribute.Call(hwnd, DWMWA_WINDOW_CORNER_PREFERENCE, uintptr(unsafe.Pointer(&original)), unsafe.Sizeof(original))
	if ret != 0 {
		t.Skip("DWM corner preferences unavailable on this Windows version")
	}
	original = 1 // DWMWCP_DONOTROUND: restore an explicit preference, not just zero.
	ret, _, _ = procDwmSetWindowAttribute.Call(hwnd, DWMWA_WINDOW_CORNER_PREFERENCE, uintptr(unsafe.Pointer(&original)), unsafe.Sizeof(original))
	if ret != 0 {
		t.Fatal("setting original DWM preference failed")
	}
	if !ApplyRoundedCorners(hwnd, 24) {
		t.Fatal("rounding failed")
	}
	var rounded uint32
	ret, _, _ = procDwmGetWindowAttribute.Call(hwnd, DWMWA_WINDOW_CORNER_PREFERENCE, uintptr(unsafe.Pointer(&rounded)), unsafe.Sizeof(rounded))
	if ret != 0 || rounded != DWMWCP_ROUND {
		t.Fatal("DWM rounding was not applied")
	}
	ResetRoundedCorners(hwnd)
	var restored uint32
	ret, _, _ = procDwmGetWindowAttribute.Call(hwnd, DWMWA_WINDOW_CORNER_PREFERENCE, uintptr(unsafe.Pointer(&restored)), unsafe.Sizeof(restored))
	if ret != 0 || restored != original {
		t.Fatalf("restored preference=%d, want %d (HRESULT 0x%X)", restored, original, ret)
	}
}

func TestRoundedCornersRestoreOriginalGDIRegion(t *testing.T) {
	hwnd := hiddenTestWindow(t)
	original, _, _ := procCreateRectRgn.Call(10, 20, 600, 300)
	if original == 0 {
		t.Fatal("create original region failed")
	}
	ret, _, _ := procSetWindowRgn.Call(hwnd, original, 0)
	if ret == 0 {
		procDeleteObject.Call(original)
		t.Fatal("set original region failed")
	}
	if !ApplyRoundedCorners(hwnd, 24) {
		t.Fatal("rounding failed")
	}
	if !ApplyRoundedCorners(hwnd, 30) {
		t.Fatal("second rounding failed")
	}
	ResetRoundedCorners(hwnd)
	region, _, _ := procCreateRectRgn.Call(0, 0, 0, 0)
	if region == 0 {
		t.Fatal("create check region failed")
	}
	defer procDeleteObject.Call(region)
	ret, _, _ = procGetWindowRgn.Call(hwnd, region)
	if ret == 0 {
		t.Fatal("original window region was lost")
	}
	var rect RECT
	ret, _, _ = gdi32.NewProc("GetRgnBox").Call(region, uintptr(unsafe.Pointer(&rect)))
	if ret == 0 || rect != (RECT{Left: 10, Top: 20, Right: 600, Bottom: 300}) {
		t.Fatalf("restored region=%+v", rect)
	}
}
