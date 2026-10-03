package main

import (
	"errors"
	"testing"

	"pip-toggle/pkg/i18n"
)

func TestLanguageSwitchKeepsModeAndReportsSaveFailure(t *testing.T) {
	oldLanguage, oldStatus := uiLanguage, currentStatus
	t.Cleanup(func() { uiLanguage, currentStatus = oldLanguage, oldStatus })
	uiLanguage, currentStatus = i18n.Russian, true
	calls := 0
	if err := changeUILanguage(i18n.English, func(value i18n.Language) error {
		calls++
		if value != i18n.English {
			t.Fatal("wrong saved language")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if uiLanguage != i18n.English || !currentStatus || calls != 1 {
		t.Fatal("language switch changed the mode or failed to save")
	}
	if tr(i18n.Exit) != "Exit" {
		t.Fatal("UI did not switch to English")
	}
	if err := changeUILanguage(i18n.Ukrainian, func(i18n.Language) error { return errors.New("access denied") }); err == nil {
		t.Fatal("save failure ignored")
	}
	if uiLanguage != i18n.English || !currentStatus {
		t.Fatal("failed save changed language or mode")
	}
	if err := changeUILanguage(i18n.Ukrainian, func(i18n.Language) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if tr(i18n.Exit) != "Вихід" || !currentStatus {
		t.Fatal("Ukrainian selection failed or changed mode")
	}
}

func TestResetUsesSelectedLanguage(t *testing.T) {
	oldLanguage := uiLanguage
	t.Cleanup(func() { uiLanguage = oldLanguage })
	for _, language := range []i18n.Language{i18n.Russian, i18n.English, i18n.Ukrainian} {
		uiLanguage = language
		for _, fail := range []bool{false, true} {
			want := i18n.ResetSuccess
			if fail {
				want = i18n.ResetError
			}
			var title, message string
			resetPiPClicks(
				func() (uintptr, string, error) { return 42, "PiP", nil },
				func(uintptr, bool) (bool, error) {
					if fail {
						return true, errors.New("access denied")
					}
					return false, nil
				},
				func(uintptr) bool { return fail },
				func(bool) {},
				func(a, b string) { title, message = a, b },
			)
			if title != i18n.Text(language, i18n.ResetTitle) || message != i18n.Text(language, want) {
				t.Fatalf("wrong notification: language=%s title=%q message=%q", language, title, message)
			}
		}
	}
}
