package cli

import (
	"os"
	"path/filepath"
)

type Paths struct {
	state    string
	settings string
}

func NewPaths(state, settings string) Paths {
	return Paths{state: state, settings: settings}
}

func (p Paths) State() string {
	if p.state != "" {
		return p.state
	}
	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return filepath.Join(dir, BinaryName, "state.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), BinaryName, "state.json")
	}
	return filepath.Join(home, ".local", "state", BinaryName, "state.json")
}

func (p Paths) Settings() string {
	if p.settings != "" {
		return p.settings
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".claude/settings.json"
	}
	return filepath.Join(home, ".claude", "settings.json")
}
