package winapi

// roundingState is protected by roundingMu when used by the background monitor.
type roundingState struct {
	enabled bool
	hwnd    uintptr
	width   int32
	height  int32
}

func (s *roundingState) reset(resetCorners func(uintptr)) {
	if s.hwnd != 0 {
		resetCorners(s.hwnd)
	}
	s.hwnd, s.width, s.height = 0, 0, 0
}

func (s *roundingState) setEnabled(enabled bool, resetCorners func(uintptr)) {
	s.enabled = enabled
	if !enabled {
		s.reset(resetCorners)
	}
}

func (s *roundingState) update(hwnd uintptr, width, height int32, apply func(uintptr, int) bool, resetCorners func(uintptr)) {
	if !s.enabled {
		return
	}
	if hwnd == 0 {
		s.reset(resetCorners)
		return
	}
	if s.hwnd != 0 && s.hwnd != hwnd {
		s.reset(resetCorners)
	}
	if width <= 0 || height <= 0 {
		return
	}
	if hwnd == s.hwnd && absInt32(width-s.width) <= 2 && absInt32(height-s.height) <= 2 {
		return
	}
	// Do not cache a failed application; retry it on the next monitor tick.
	if apply(hwnd, 24) {
		s.hwnd, s.width, s.height = hwnd, width, height
	}
}
