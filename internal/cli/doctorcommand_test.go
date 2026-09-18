package cli

import (
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
