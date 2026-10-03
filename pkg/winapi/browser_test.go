package winapi

import "testing"

func TestSupportedBrowserProcesses(t *testing.T) {
	for _, process := range []string{"opera.exe", "chrome.exe", "msedge.exe", "CHROME.EXE", `C:\Program Files\Google\Chrome\Application\chrome.exe`} {
		if !IsSupportedBrowserProcess(process) {
			t.Errorf("supported browser rejected: %q", process)
		}
	}
	for _, process := range []string{"", "firefox.exe", "msedgewebview2.exe", "electron.exe", "chrome_helper.exe", "fake-chrome.exe", "opera.exe.bak"} {
		if IsSupportedBrowserProcess(process) {
			t.Errorf("unrelated process accepted: %q", process)
		}
	}
}

func TestNativePiPTitles(t *testing.T) {
	for _, title := range []string{"Picture in picture", "Picture-in-Picture", "Picture\u2011in\u2011Picture", "Картинка в картинке", "  КАРТИНКА В КАРТИНКЕ  ", "PiP", "Picture\u00a0in\u00a0picture"} {
		if !IsPiPTitle(title) {
			t.Errorf("native PiP title rejected: %q", title)
		}
	}
	for _, title := range []string{"", "PiP - Google Chrome", "Picture in picture - Microsoft Edge", "Картинка в картинке — Opera", "pip install package", "Pipelines", "My picture-in-picture video", "Google Chrome"} {
		if IsPiPTitle(title) {
			t.Errorf("regular tab title accepted as PiP: %q", title)
		}
	}
}

func TestPiPRequiresNativeClassAndTopmostStyle(t *testing.T) {
	for _, tc := range []struct {
		class string
		style uintptr
		want  bool
	}{
		{"Chrome_WidgetWin_1", WS_EX_TOPMOST, true},
		{"Chrome_WidgetWin_1", WS_EX_TOPMOST | WS_EX_LAYERED | WS_EX_TRANSPARENT, true},
		{"Chrome_WidgetWin_1", 0, false},
		{"Chrome_WidgetWin_1", WS_EX_LAYERED, false},
		{"Chrome_WidgetWin_0", WS_EX_TOPMOST, false},
		{"MozillaDialogClass", WS_EX_TOPMOST, false},
		{"", WS_EX_TOPMOST, false},
	} {
		if got := hasPiPWindowStyle(tc.class, tc.style); got != tc.want {
			t.Errorf("class=%q style=0x%X: got %v, want %v", tc.class, tc.style, got, tc.want)
		}
	}
}
