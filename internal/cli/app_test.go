package cli

import (
	"strings"
	"testing"
)

func TestUnknownSubcommandFails(t *testing.T) {
	h := newHarness(t)

	if code := h.run("", "frobnicate"); code == 0 {
		t.Error("unknown subcommand exited 0, want non-zero")
	}
	if !strings.Contains(h.stderr.String(), "frobnicate") {
		t.Errorf("stderr = %q, want it to name the bad subcommand", h.stderr)
	}
}

func TestNoSubcommandPrintsUsage(t *testing.T) {
	h := newHarness(t)

	h.run("")
	if !strings.Contains(h.stderr.String()+h.stdout.String(), "collect") {
		t.Error("bare invocation should show usage listing the subcommands")
	}
}

func TestVersion(t *testing.T) {
	h := newHarness(t)

	if code := h.run("", "version"); code != 0 {
		t.Fatalf("version exit = %d", code)
	}
	if strings.TrimSpace(h.stdout.String()) == "" {
		t.Error("version printed nothing")
	}
}
