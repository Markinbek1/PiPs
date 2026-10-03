package main

import (
	"fmt"

	"github.com/getlantern/systray"
	"pip-toggle/pkg/i18n"
)

func stateKey(enabled bool, on, off i18n.Key) i18n.Key {
	if enabled {
		return on
	}
	return off
}

func setMenuText(item *systray.MenuItem, title, tooltip i18n.Key) {
	item.SetTitle(tr(title))
	item.SetTooltip(tr(tooltip))
}

// Don't change the displayed language if saving the user's choice fails.
func changeUILanguage(language i18n.Language, save func(i18n.Language) error) error {
	if !language.Valid() {
		return fmt.Errorf("unsupported language: %q", language)
	}
	if language == uiLanguage {
		return nil
	}
	if err := save(language); err != nil {
		return err
	}
	uiLanguage = language
	return nil
}
