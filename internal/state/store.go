package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/statusline"
	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

type Store struct {
	path string
}

func New(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Path() string { return s.path }

func (s *Store) Merge(p statusline.Payload, now time.Time) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	lock := newFileLock(s.path + ".lock")
	if err := lock.Acquire(); err != nil {
		return err
	}
	defer lock.Release()

	f, err := s.read()
	if err != nil {
		return err
	}
	f.absorb(p, now)

	b, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	return newAtomicFile(s.path, 0o600).Write(b)
}

func (s *Store) Snapshot() (usage.Snapshot, error) {
	f, err := s.read()
	if err != nil {
		return usage.Snapshot{}, err
	}
	return f.domain(), nil
}

func (s *Store) read() (snapshotFile, error) {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return newSnapshotFile(), nil
	}
	if err != nil {
		return snapshotFile{}, fmt.Errorf("read state file %s: %w", s.path, err)
	}
	f := newSnapshotFile()
	if err := json.Unmarshal(b, &f); err != nil {
		return snapshotFile{}, fmt.Errorf("parse state file %s: %w", s.path, err)
	}
	if f.Sessions == nil {
		f.Sessions = newSessions()
	}
	return f, nil
}
