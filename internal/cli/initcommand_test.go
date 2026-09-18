package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitWritesTheStatusLineIntoSettings(t *testing.T) {
	h := newHarness(t)

	if code := h.run("", "init"); code != 0 {
		t.Fatalf("init exit = %d, stderr = %s", code, h.stderr)
	}
	b, err := os.ReadFile(filepath.Join(h.dir, "settings.json"))
	if err != nil {
		t.Fatalf("settings not written: %v", err)
	}
	if !strings.Contains(string(b), "collect") {
		t.Errorf("settings = %s, want a statusLine running collect", b)
	}
	if !strings.Contains(h.stdout.String(), "i3blocks") {
		t.Errorf("init should print the bar snippet, got: %s", h.stdout)
	}
}
