package main

import (
	"errors"
	"strings"
	"testing"
)

func TestEmergencyResetReportsActualOutcome(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		findErr, resetErr                   error
		actualState, wantState, wantSuccess bool
		wantCalls                           int
	}{
		{"successful reset", nil, nil, true, false, true, 1},
		{"failed reset with clicks still disabled", nil, errors.New("access denied"), true, true, false, 1},
		{"failed reset with clicks already restored", nil, errors.New("window closed"), false, false, false, 1},
		{"window missing", errors.New("not found"), nil, true, false, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls, notifications, updates := 0, 0, 0
			var state bool
			var message string
			resetPiPClicks(
				func() (uintptr, string, error) { return 42, "PiP", tc.findErr },
				func(hwnd uintptr, enable bool) (bool, error) {
					if hwnd != 42 || enable {
						t.Fatal("reset must disable click-through for the selected window")
					}
					calls++
					return false, tc.resetErr
				},
				func(uintptr) bool { return tc.actualState },
				func(value bool) { updates++; state = value },
				func(_, value string) { notifications++; message = value },
			)
			if calls != tc.wantCalls || notifications != 1 || updates != 1 || state != tc.wantState {
				t.Fatalf("calls=%d notifications=%d updates=%d state=%v", calls, notifications, updates, state)
			}
			if success := strings.Contains(message, "клики восстановлены"); success != tc.wantSuccess {
				t.Fatalf("incorrect notification: %q", message)
			}
		})
	}
}
