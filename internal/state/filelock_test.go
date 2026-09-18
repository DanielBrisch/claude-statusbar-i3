//go:build unix

package state

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func holdLock(t *testing.T, path string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open lock: %v", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock: %v", err)
	}
	t.Cleanup(func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	})
}
func TestMergeGivesUpInsteadOfBlockingOnAHeldLock(t *testing.T) {
	dir := t.TempDir()
	s := New(filepath.Join(dir, "state.json"))
	holdLock(t, s.Path()+".lock")
	start := time.Now()
	err := s.Merge(payload("a", 10), now)
	elapsed := time.Since(start)
	if !errors.Is(err, ErrBusy) {
		t.Fatalf("Merge returned %v, want ErrBusy — blocking forever would hang Claude Code's status line", err)
	}
	if elapsed > 2*LockTimeout {
		t.Errorf("Merge waited %v, want it to give up around %v", elapsed, LockTimeout)
	}
}
func TestMergeSucceedsOnceTheLockIsFree(t *testing.T) {
	dir := t.TempDir()
	s := New(filepath.Join(dir, "state.json"))
	if err := s.Merge(payload("a", 10), now); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if err := s.Merge(payload("b", 20), now); err != nil {
		t.Fatalf("second Merge: %v", err)
	}
}
