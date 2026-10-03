package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type settings struct {
	Language Language `json:"language"`
}

// LoadLanguage defaults to Russian on first use or invalid settings.
func LoadLanguage(path string) (Language, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Russian, nil
	}
	if err != nil {
		return Russian, err
	}
	var saved settings
	if err := json.Unmarshal(data, &saved); err != nil {
		return Russian, err
	}
	if !saved.Language.Valid() {
		return Russian, fmt.Errorf("unsupported language: %q", saved.Language)
	}
	return saved.Language, nil
}

// SaveLanguage replaces the settings only after a complete write succeeds.
func SaveLanguage(path string, language Language) error {
	if path == "" || !language.Valid() {
		return fmt.Errorf("invalid settings path or language")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings{Language: language}, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".settings-*")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)
	if _, err := file.Write(append(data, '\n')); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
