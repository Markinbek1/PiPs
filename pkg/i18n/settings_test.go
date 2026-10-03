package i18n

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLanguageSettingsAcrossRestarts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "PiPs", "settings.json")
	if language, err := LoadLanguage(path); err != nil || language != Russian {
		t.Fatalf("first start: language=%q err=%v", language, err)
	}
	for _, want := range []Language{English, Ukrainian, Russian, English} {
		if err := SaveLanguage(path, want); err != nil {
			t.Fatal(err)
		}
		if got, err := LoadLanguage(path); err != nil || got != want {
			t.Fatalf("saved %q; loaded %q, err=%v", want, got, err)
		}
		entries, err := os.ReadDir(filepath.Dir(path))
		if err != nil || len(entries) != 1 {
			t.Fatalf("temporary settings left behind: %v, %v", entries, err)
		}
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveLanguage(path, Language("unknown")); err == nil {
		t.Fatal("invalid language saved")
	}
	after, err := os.ReadFile(path)
	if err != nil || string(before) != string(after) {
		t.Fatal("invalid choice replaced existing settings")
	}
}

func TestUnreadableOrInvalidSettingsFallBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	for _, contents := range []string{`{"language":"unknown"}`, `{"language":`, `{}`} {
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
		if got, err := LoadLanguage(path); err == nil || got != Russian {
			t.Fatalf("invalid settings %q: language=%q err=%v", contents, got, err)
		}
	}
	if got, err := LoadLanguage(t.TempDir()); err == nil || got != Russian {
		t.Fatal("directory should fall back with an error")
	}
	blocked := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocked, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := SaveLanguage(filepath.Join(blocked, "settings.json"), English); err == nil {
		t.Fatal("save to invalid directory succeeded")
	}
}

func TestEveryVisibleMessageIsTranslated(t *testing.T) {
	for key := Key(0); key < keyCount; key++ {
		for _, language := range []Language{Russian, English, Ukrainian} {
			if strings.TrimSpace(Text(language, key)) == "" {
				t.Errorf("empty translation: key=%d language=%s", key, language)
			}
		}
		if Text(Language("unknown"), key) != Text(Russian, key) {
			t.Errorf("fallback missing for key=%d", key)
		}
	}
	for _, language := range []Language{Russian, English, Ukrainian} {
		for key, hotkey := range map[Key]string{Toggle: "Ctrl+Alt+P", PlayPause: "Ctrl+Alt+Space", Reset: "Ctrl+Alt+Shift+P"} {
			if !strings.Contains(Text(language, key), hotkey) {
				t.Errorf("hotkey missing: language=%s key=%d", language, key)
			}
		}
	}
}
