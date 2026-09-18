package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorExplainsAnEmptyState(t *testing.T) {
	h := newHarness(t)

	h.run("", "doctor")
	out := h.stdout.String()
	if !strings.Contains(out, "state.json") {
		t.Errorf("doctor output = %q, want the state path", out)
	}
	if !strings.Contains(strings.ToLower(out), "statusline") {
		t.Errorf("doctor output = %q, want it to point at the statusLine setup", out)
	}
}

func TestDoctorReportsAMissingStatusLine(t *testing.T) {
	h := newHarness(t)
	h.run("", "doctor")

	out := h.stdout.String()
	if !strings.Contains(out, "statusLine:") {
		t.Errorf("doctor output = %q, want a statusLine line", out)
	}
	if !strings.Contains(out, "not configured") {
		t.Errorf("doctor output = %q, want it to say the statusLine is missing — that is the usual reason the block is empty", out)
	}
}

func TestDoctorReportsAStatusLinePointingElsewhere(t *testing.T) {
	h := newHarness(t)
	if err := os.WriteFile(filepath.Join(h.dir, "settings.json"),
		[]byte(`{"statusLine":{"type":"command","command":"~/my-statusline.sh"}}`), 0o600); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	h.run("", "doctor")
	out := h.stdout.String()
	if !strings.Contains(out, "my-statusline.sh") {
		t.Errorf("doctor output = %q, want the command it actually found", out)
	}
	if !strings.Contains(out, "--passthrough") {
		t.Errorf("doctor output = %q, want it to point at the way to keep both", out)
	}
}

func TestDoctorAcceptsAStatusLineThatRunsThisBinary(t *testing.T) {
	h := newHarness(t)
	if err := os.WriteFile(filepath.Join(h.dir, "settings.json"),
		[]byte(`{"statusLine":{"type":"command","command":"$HOME/.local/bin/claude-statusbar collect"}}`), 0o600); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	h.run("", "doctor")
	if out := h.stdout.String(); !strings.Contains(out, "statusLine: ok") {
		t.Errorf("doctor output = %q, want it to accept a statusLine that runs the collector", out)
	}
}
