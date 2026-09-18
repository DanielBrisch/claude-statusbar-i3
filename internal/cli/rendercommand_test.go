package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderOnAnEmptyStateIsSilentAndSuccessful(t *testing.T) {
	h := newHarness(t)

	if code := h.run("", "render", "--format", "i3blocks"); code != 0 {
		t.Errorf("render exit = %d, want 0 — a missing state must never break the bar", code)
	}
	if h.stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", h.stdout)
	}
}

func TestRenderStaysQuietAtUrgentByDefault(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(97), "collect", "--print", "none")

	if code := h.run("", "render", "--format", "i3blocks"); code != 0 {
		t.Errorf("render exit = %d, want 0 — exit 33 recolours the block, so it is opt-in", code)
	}
	if strings.Contains(h.stdout.String(), "#") {
		t.Errorf("stdout = %q, want no colour line by default", h.stdout)
	}
}

func TestRenderExitsThirtyThreeWhenUrgentExitIsAskedFor(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(97), "collect", "--print", "none")

	if code := h.run("", "render", "--format", "i3blocks", "--urgent-exit"); code != 33 {
		t.Errorf("render exit = %d, want 33 (i3blocks urgent)", code)
	}
}

func TestRenderEmitsTheColourLineOnlyWhenOneIsConfigured(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(97), "collect", "--print", "none")

	h.run("", "render", "--format", "i3blocks", "--color-crit", "#E06C75")
	if !strings.Contains(h.stdout.String(), "#E06C75") {
		t.Errorf("stdout = %q, want the configured colour", h.stdout)
	}
}

func TestRenderDoesNotExitThirtyThreeForOtherFormats(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(97), "collect", "--print", "none")

	if code := h.run("", "render", "--format", "waybar", "--urgent-exit"); code != 0 {
		t.Errorf("render exit = %d, want 0 — exit 33 is an i3blocks convention", code)
	}
}

func TestRenderWaybarIsValidJSON(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(63), "collect", "--print", "none")
	h.run("", "render", "--format", "waybar")

	var out map[string]any
	if err := json.Unmarshal(h.stdout.Bytes(), &out); err != nil {
		t.Fatalf("waybar output is not JSON: %v\n%s", err, h.stdout)
	}
	if out["text"] != "✳ 63% 1h42 │ week 21% 4d" {
		t.Errorf("text = %v", out["text"])
	}
}

func TestClickingTheI3blocksBlockDoesNothing(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(63), "collect", "--print", "none")

	for _, button := range []string{"1", "2", "3", "4", "5"} {
		h.env["BLOCK_BUTTON"] = button
		if code := h.run("", "render", "--format", "i3blocks"); code != 0 {
			t.Errorf("button %s: exit = %d", button, code)
		}
	}
	if h.notifier.sent != 0 {
		t.Errorf("notifier called %d times, want 0 — the icon makes the block self-explanatory, so the click no longer pops anything", h.notifier.sent)
	}
	if !strings.HasPrefix(h.stdout.String(), "\u2733 63%") {
		t.Errorf("the block must still render normally, got %q", h.stdout)
	}
}
