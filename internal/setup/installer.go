package setup

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const settingsKey = "statusLine"

type Installer struct {
	path string
}

func NewInstaller(settingsPath string) *Installer {
	return &Installer{path: settingsPath}
}

func (i *Installer) Path() string { return i.path }

func (i *Installer) Install(command string, force bool) error {
	raw, err := os.ReadFile(i.path)
	missing := errors.Is(err, os.ErrNotExist)
	if err != nil && !missing {
		return fmt.Errorf("read %s: %w", i.path, err)
	}

	settings := map[string]json.RawMessage{}
	if !missing {
		if err := json.Unmarshal(raw, &settings); err != nil {
			return fmt.Errorf("parse %s: %w — fix it by hand before running init", i.path, err)
		}
		if existing, taken := settings[settingsKey]; taken && !force {
			return i.refuse(existing, command)
		}
	}

	encoded, err := json.Marshal(newStatusLine(command))
	if err != nil {
		return fmt.Errorf("encode statusLine: %w", err)
	}
	settings[settingsKey] = encoded

	if !missing {
		if err := os.WriteFile(i.path+".bak", raw, 0o600); err != nil {
			return fmt.Errorf("write backup: %w", err)
		}
	}
	return i.save(settings)
}

func (i *Installer) refuse(existing json.RawMessage, command string) error {
	return fmt.Errorf(
		"%s already defines a statusLine (%s); rerun with --force to replace it, "+
			"or keep both by pointing it at: %s --passthrough '<your current command>'",
		i.path, existing, command)
}

func (i *Installer) save(settings map[string]json.RawMessage) error {
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(i.path), 0o755); err != nil {
		return fmt.Errorf("create settings dir: %w", err)
	}
	if err := os.WriteFile(i.path, append(out, '\n'), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", i.path, err)
	}
	return nil
}
