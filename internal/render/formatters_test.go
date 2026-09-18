package render

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

func lines(out Output) []string {
	return strings.Split(strings.TrimRight(out.Text, "\n"), "\n")
}

func TestI3blocksEmitsFullShortAndColour(t *testing.T) {
	s := live()
	s.FiveHour.UsedPercentage = 90

	got := lines(render("i3blocks", s, coloured(opts())))
	if len(got) != 3 {
		t.Fatalf("emitted %d lines, want 3: %q", len(got), got)
	}
	if got[0] != "✳ 90% 1h42 │ week 21% 4d" {
		t.Errorf("full_text = %q", got[0])
	}
	if got[1] != "90%│21%" {
		t.Errorf("short_text = %q", got[1])
	}
	if got[2] != "#E06C75" {
		t.Errorf("color = %q, want %q", got[2], "#E06C75")
	}
}

func TestI3blocksOmitsTheColourLineWhenThereIsNoColour(t *testing.T) {
	s := live()
	s.FiveHour.UsedPercentage = 97

	if got := lines(render("i3blocks", s, opts())); len(got) != 2 {
		t.Fatalf("emitted %d lines, want 2 (no colour line): %q", len(got), got)
	}
}

func TestI3blocksEmitsNothingWhenThereIsNothingToShow(t *testing.T) {
	if out := render("i3blocks", usage.Snapshot{}, opts()); !out.Empty() {
		t.Errorf("Text = %q, want empty", out.Text)
	}
}

func TestI3blocksOwnsItsUrgentExitCode(t *testing.T) {
	s := live()
	s.FiveHour.UsedPercentage = 97

	if got := render("i3blocks", s, opts()).ExitCode; got != 0 {
		t.Errorf("ExitCode = %d, want 0 — exit 33 recolours the block, so it is opt-in", got)
	}

	o := opts()
	o.UrgentExit = true
	if got := render("i3blocks", s, o).ExitCode; got != 33 {
		t.Errorf("ExitCode = %d, want 33", got)
	}
}

func TestOnlyI3blocksEverSetsAnExitCode(t *testing.T) {
	s := live()
	s.FiveHour.UsedPercentage = 97
	o := opts()
	o.UrgentExit = true

	for _, name := range []string{"waybar", "polybar", "plain", "json"} {
		if got := render(name, s, o).ExitCode; got != 0 {
			t.Errorf("%s ExitCode = %d, want 0 — exit 33 is an i3blocks convention", name, got)
		}
	}
}

func TestWaybarCarriesTheDetailInTheTooltip(t *testing.T) {
	var out waybarView
	if err := json.Unmarshal([]byte(render("waybar", live(), opts()).Text), &out); err != nil {
		t.Fatalf("Waybar output is not valid JSON: %v", err)
	}
	if out.Text != "✳ 63% 1h42 │ week 21% 4d" {
		t.Errorf("text = %q", out.Text)
	}
	if !strings.Contains(out.Tooltip, "Weekly") {
		t.Errorf("tooltip = %q, want the breakdown", out.Tooltip)
	}
	if out.Percentage != 63 {
		t.Errorf("percentage = %v, want 63", out.Percentage)
	}
}

func TestWaybarStillReportsTheLevelSoCSSCanStyleIt(t *testing.T) {
	s := live()
	s.FiveHour.UsedPercentage = 97

	var out waybarView
	if err := json.Unmarshal([]byte(render("waybar", s, opts()).Text), &out); err != nil {
		t.Fatalf("Waybar output is not JSON: %v", err)
	}
	if out.Class != "urgent" {
		t.Errorf("class = %q, want %q — no colour is forced, but the level must stay available to CSS", out.Class, "urgent")
	}
}

func TestPolybarWrapsColourAndClickAction(t *testing.T) {
	s := live()
	s.FiveHour.UsedPercentage = 90

	got := render("polybar", s, coloured(opts())).Text
	if !strings.Contains(got, "%{F#E06C75}") {
		t.Errorf("Polybar = %q, want a colour tag", got)
	}
	if !strings.Contains(got, "%{A1:"+DetailCommand+":}") || !strings.HasSuffix(got, "%{A}") {
		t.Errorf("Polybar = %q, want a left-click action", got)
	}
}

func TestPolybarLeavesTheColourAloneByDefault(t *testing.T) {
	s := live()
	s.FiveHour.UsedPercentage = 97

	if got := render("polybar", s, opts()).Text; strings.Contains(got, "%{F") {
		t.Errorf("Polybar = %q, want no colour tag by default", got)
	}
}

func TestPlainHasNoMarkup(t *testing.T) {
	got := render("plain", live(), opts()).Text
	if strings.Contains(got, "%{") || strings.Contains(got, "#") {
		t.Errorf("Plain = %q, want no markup", got)
	}
}

func TestPangoIsOnlyForBarsThatParseIt(t *testing.T) {
	o := opts()
	o.Markup = MarkupPango

	for _, name := range []string{"plain", "json", "polybar"} {
		if out := render(name, live(), o).Text; strings.Contains(out, "<span") {
			t.Errorf("%s output carries pango markup:\n%s", name, out)
		}
	}
}

func TestJSONStillExposesTheSessionForCustomFormats(t *testing.T) {
	var out snapshotView
	if err := json.Unmarshal([]byte(render("json", live(), opts()).Text), &out); err != nil {
		t.Fatalf("JSON output is not valid: %v", err)
	}
	if out.Session == nil || out.Session.Model != "Opus" {
		t.Errorf("session = %+v, want it still available to --format json and --template", out.Session)
	}
	if out.FiveHour == nil || out.FiveHour.TimeLeft != "1h42" {
		t.Errorf("five_hour = %+v", out.FiveHour)
	}
	if out.SpendLimit != nil {
		t.Errorf("spend_limit = %+v, want nil when absent", out.SpendLimit)
	}
	_ = time.Minute
}
