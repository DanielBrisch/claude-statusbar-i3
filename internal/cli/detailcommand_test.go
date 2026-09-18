package cli

import (
	"strings"
	"testing"
)

func TestDetailNotifyUsesTheNotifier(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(63), "collect", "--print", "none")

	if code := h.run("", "detail", "--notify"); code != 0 {
		t.Fatalf("detail exit = %d, stderr = %s", code, h.stderr)
	}
	if h.notifier.sent != 1 {
		t.Errorf("notifier called %d times, want 1", h.notifier.sent)
	}
}

func TestDetailWithoutNotifyPrintsToStdout(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(63), "collect", "--print", "none")

	h.run("", "detail")
	if !strings.Contains(h.stdout.String(), "5h window") {
		t.Errorf("stdout = %q, want the detail", h.stdout)
	}
	if h.notifier.sent != 0 {
		t.Errorf("notifier called %d times without --notify, want 0", h.notifier.sent)
	}
}

func TestDetailNotifyFallbackStillReachesStdout(t *testing.T) {
	h := newHarness(t)
	h.notifier.fallback = true
	h.run(payloadJSON(63), "collect", "--print", "none")

	h.run("", "detail", "--notify")
	if !strings.Contains(h.stdout.String(), "Weekly") {
		t.Errorf("stdout = %q, want the breakdown — on a headless box the fallback is the whole point", h.stdout)
	}
}
