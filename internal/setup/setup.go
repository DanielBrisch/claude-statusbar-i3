package setup

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type statusLine struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func InstallStatusLine(settingsPath, command string, force bool) error {
	raw, err := os.ReadFile(settingsPath)
	missing := errors.Is(err, os.ErrNotExist)
	if err != nil && !missing {
		return fmt.Errorf("read %s: %w", settingsPath, err)
	}

	settings := map[string]json.RawMessage{}
	if !missing {
		if err := json.Unmarshal(raw, &settings); err != nil {
			return fmt.Errorf("parse %s: %w — fix it by hand before running init", settingsPath, err)
		}
		if existing, ok := settings["statusLine"]; ok && !force {
			return fmt.Errorf(
				"%s already defines a statusLine (%s); rerun with --force to replace it, "+
					"or keep both by pointing it at: %s --passthrough '<your current command>'",
				settingsPath, existing, command)
		}
	}

	encoded, err := json.Marshal(statusLine{Type: "command", Command: command})
	if err != nil {
		return fmt.Errorf("encode statusLine: %w", err)
	}
	settings["statusLine"] = encoded

	if !missing {
		if err := os.WriteFile(settingsPath+".bak", raw, 0o600); err != nil {
			return fmt.Errorf("write backup: %w", err)
		}
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("create settings dir: %w", err)
	}
	if err := os.WriteFile(settingsPath, append(out, '\n'), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", settingsPath, err)
	}
	return nil
}
