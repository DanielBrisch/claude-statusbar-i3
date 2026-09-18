package state

import (
	"fmt"
	"os"
	"path/filepath"
)

type atomicFile struct {
	path string
	mode os.FileMode
}

func newAtomicFile(path string, mode os.FileMode) atomicFile {
	return atomicFile{path: path, mode: mode}
}

func (a atomicFile) Write(b []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(a.path), ".state-*.json")
	if err != nil {
		return fmt.Errorf("create temp state: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp state: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp state: %w", err)
	}
	if err := os.Chmod(tmp.Name(), a.mode); err != nil {
		return fmt.Errorf("chmod temp state: %w", err)
	}
	if err := os.Rename(tmp.Name(), a.path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	return nil
}
