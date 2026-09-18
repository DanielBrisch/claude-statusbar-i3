package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectStoresStateAndEchoesTheCompactLine(t *testing.T) {
	h := newHarness(t)

	if code := h.run(payloadJSON(63), "collect"); code != 0 {
		t.Fatalf("collect exit = %d, stderr = %s", code, h.stderr)
	}
	if got, want := strings.TrimSpace(h.stdout.String()), "✳ 63% 1h42 │ week 21% 4d"; got != want {
		t.Errorf("collect stdout = %q, want %q", got, want)
	}

	if code := h.run("", "render", "--format", "plain"); code != 0 {
		t.Fatalf("render exit = %d, stderr = %s", code, h.stderr)
	}
	if got, want := strings.TrimSpace(h.stdout.String()), "✳ 63% 1h42 │ week 21% 4d"; got != want {
		t.Errorf("render stdout = %q, want %q", got, want)
	}
}

func TestCollectPrintNoneStillWritesState(t *testing.T) {
	h := newHarness(t)

	if code := h.run(payloadJSON(63), "collect", "--print", "none"); code != 0 {
		t.Fatalf("collect exit = %d, stderr = %s", code, h.stderr)
	}
	if h.stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", h.stdout)
	}
	if _, err := os.Stat(filepath.Join(h.dir, "state.json")); err != nil {
		t.Errorf("state was not written: %v", err)
	}
}

func TestCollectSurvivesAPayloadWithoutRateLimits(t *testing.T) {
	h := newHarness(t)

	code := h.run(`{"session_id":"s1","model":{"display_name":"Opus"}}`, "collect")
	if code != 0 {
		t.Fatalf("collect exit = %d, stderr = %s", code, h.stderr)
	}
	if h.stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty when there are no rate limits", h.stdout)
	}
}

func TestCollectReportsMalformedInput(t *testing.T) {
	h := newHarness(t)

	if code := h.run("{not json", "collect"); code == 0 {
		t.Fatal("collect of malformed JSON exited 0, want non-zero")
	}
	if !strings.Contains(h.stderr.String(), "decode") {
		t.Errorf("stderr = %q, want it to explain the decode failure", h.stderr)
	}
}

func TestCollectRefreshPokesTheBarAfterWritingState(t *testing.T) {
	h := newHarness(t)

	if code := h.run(payloadJSON(63), "collect", "--print", "none", "--refresh", "pkill -SIGRTMIN+12 i3blocks"); code != 0 {
		t.Fatalf("collect exit = %d, stderr = %s", code, h.stderr)
	}
	if len(h.shell) != 1 || h.shell[0] != "pkill -SIGRTMIN+12 i3blocks" {
		t.Fatalf("shell calls = %v, want the refresh command", h.shell)
	}
	if _, err := os.Stat(filepath.Join(h.dir, "state.json")); err != nil {
		t.Errorf("state was not written before the refresh: %v", err)
	}
}

func TestCollectDoesNotPokeTheBarWhenThePayloadIsRejected(t *testing.T) {
	h := newHarness(t)

	if code := h.run("{not json", "collect", "--refresh", "pkill -SIGRTMIN+12 i3blocks"); code == 0 {
		t.Fatal("collect exited 0 on malformed input")
	}
	if len(h.shell) != 0 {
		t.Errorf("shell calls = %v, want none — nothing changed, so the bar has nothing to re-read", h.shell)
	}
}

func TestCollectSurvivesAFailingRefreshCommand(t *testing.T) {
	h := newHarness(t)
	h.shellErr = errors.New("no such process")

	if code := h.run(payloadJSON(63), "collect", "--print", "none", "--refresh", "pkill -SIGRTMIN+12 i3blocks"); code != 0 {
		t.Errorf("collect exit = %d, want 0 — a dead bar must not break Claude Code's status line", code)
	}
}

func TestCollectPassthroughFeedsTheOriginalPayloadOnward(t *testing.T) {
	h := newHarness(t)
	in := payloadJSON(63)

	if code := h.run(in, "collect", "--passthrough", "my-statusline.sh"); code != 0 {
		t.Fatalf("collect exit = %d, stderr = %s", code, h.stderr)
	}
	if len(h.shell) != 1 || h.shell[0] != "my-statusline.sh" {
		t.Fatalf("shell calls = %v, want the passthrough command", h.shell)
	}
	if !strings.Contains(h.stdout.String(), fmt.Sprintf("passthrough saw %d bytes", len(in))) {
		t.Errorf("stdout = %q, want the passthrough output built from the full payload", h.stdout)
	}
}
