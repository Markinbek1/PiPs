package winapi

import "testing"

func TestRoundingDisableAndReenableSameWindow(t *testing.T) {
	state := roundingState{enabled: true}
	var applied, reset []uintptr
	apply := func(hwnd uintptr, _ int) bool { applied = append(applied, hwnd); return true }
	restore := func(hwnd uintptr) { reset = append(reset, hwnd) }
	state.update(42, 640, 360, apply, restore)
	state.setEnabled(false, restore)
	state.update(42, 640, 360, apply, restore)
	if len(reset) != 1 || reset[0] != 42 || len(applied) != 1 {
		t.Fatalf("disabled rounding must restore the real window and stay off: applied=%v reset=%v", applied, reset)
	}
	state.setEnabled(true, restore)
	state.update(42, 640, 360, apply, restore)
	if len(applied) != 2 {
		t.Fatal("reenabling must reapply rounding even when the window has not resized")
	}
}

func TestRoundingWindowChangesAndCloses(t *testing.T) {
	state := roundingState{enabled: true}
	var reset []uintptr
	apply := func(uintptr, int) bool { return true }
	restore := func(hwnd uintptr) { reset = append(reset, hwnd) }
	state.update(1, 640, 360, apply, restore)
	state.update(2, 640, 360, apply, restore)
	state.update(0, 0, 0, apply, restore)
	if len(reset) != 2 || reset[0] != 1 || reset[1] != 2 || state.hwnd != 0 {
		t.Fatalf("previous windows must be restored: reset=%v state=%+v", reset, state)
	}
}

func TestRoundingRetriesFailuresAndIgnoresSizeJitter(t *testing.T) {
	state := roundingState{enabled: true}
	attempts := 0
	apply := func(uintptr, int) bool { attempts++; return attempts > 1 }
	restore := func(uintptr) {}
	state.update(1, 640, 360, apply, restore)
	state.update(1, 640, 360, apply, restore)
	state.update(1, 641, 362, apply, restore)
	if attempts != 2 {
		t.Fatalf("failed rounding must be retried, size jitter must not: attempts=%d", attempts)
	}
	state.update(1, 650, 360, apply, restore)
	if attempts != 3 {
		t.Fatal("real resizing must update rounding")
	}
}
