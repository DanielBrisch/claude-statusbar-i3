//go:build unix

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestCollectSurvivesABusyStateFile(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(63), "collect", "--print", "none")

	lock := filepath.Join(h.dir, "state.json.lock")
	f, err := os.OpenFile(lock, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open lock: %v", err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock: %v", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	code := h.run(payloadJSON(80), "collect")
	if code != 0 {
		t.Errorf("collect exit = %d, want 0 — a contended lock must not break Claude Code's status line", code)
	}
	if !strings.Contains(h.stdout.String(), "63%") {
		t.Errorf("stdout = %q, want the last known figures — the tick is lost, the reading is not", h.stdout)
	}
}
