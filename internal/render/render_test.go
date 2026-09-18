package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

var now = time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)

func at(d time.Duration) time.Time { return now.Add(d) }

func opts() Options {
	return Options{
		Now:         now,
		Label:       DefaultLabel,
		WeeklyLabel: DefaultWeeklyLabel,
		Thresholds:  usage.DefaultThresholds(),
		Colors:      DefaultColors(),
	}
}

func live() usage.Snapshot {
	return usage.Snapshot{
		FiveHour:  &usage.Limit{UsedPercentage: 63, ResetsAt: at(102 * time.Minute)},
		SevenDay:  &usage.Limit{UsedPercentage: 21, ResetsAt: at(100 * time.Hour)},
		UpdatedAt: at(-2 * time.Minute),
		Session: &usage.Session{
			Model: "Opus", CostUSD: 2.41, ContextUsedPct: 18,
			ContextWindowSize: 200000, Duration: 45 * time.Minute, UpdatedAt: at(-2 * time.Minute),
		},
	}
}

func TestCompactShowsBothWindows(t *testing.T) {
	got := Compact(live(), opts()).FullText
	want := "session 63% 1h42 │ week 21% 4d"
	if got != want {
		t.Errorf("Compact() = %q, want %q", got, want)
	}
}

func TestCompactPlaceholdersTheSessionButKeepsTheWeekly(t *testing.T) {
	s := live()
	s.FiveHour = &usage.Limit{UsedPercentage: 99, ResetsAt: at(-time.Minute)}

	got := Compact(s, opts()).FullText
	want := "session — │ week 21% 4d"
	if got != want {
		t.Errorf("Compact() = %q, want %q", got, want)
	}
}

func TestCompactIsEmptyWhenNothingIsLive(t *testing.T) {
	s := usage.Snapshot{
		FiveHour: &usage.Limit{ResetsAt: at(-time.Hour)},
		SevenDay: &usage.Limit{ResetsAt: at(-time.Hour)},
	}
	if got := Compact(s, opts()).FullText; got != "" {
		t.Errorf("Compact() = %q, want empty", got)
	}
}

func TestCompactWithoutAnyRateLimitsIsEmpty(t *testing.T) {
	if got := Compact(usage.Snapshot{}, opts()).FullText; got != "" {
		t.Errorf("Compact() = %q, want empty", got)
	}
}

func TestCompactColorFollowsTheWorstWindow(t *testing.T) {
	cases := []struct {
		name      string
		fiveHour  float64
		wantColor string
		wantLevel usage.Level
	}{
		{"calm", 10, "", usage.LevelOK},
		{"warn", 70, DefaultColors().Warn, usage.LevelWarn},
		{"crit", 90, DefaultColors().Crit, usage.LevelCrit},
		{"urgent", 97, DefaultColors().Crit, usage.LevelUrgent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := live()
			s.FiveHour.UsedPercentage = tc.fiveHour
			r := Compact(s, opts())
			if r.Color != tc.wantColor {
				t.Errorf("Color = %q, want %q", r.Color, tc.wantColor)
			}
			if r.Level != tc.wantLevel {
				t.Errorf("Level = %v, want %v", r.Level, tc.wantLevel)
			}
		})
	}
}

func TestI3blocksEmitsFullShortAndColour(t *testing.T) {
	s := live()
	s.FiveHour.UsedPercentage = 90

	lines := strings.Split(strings.TrimRight(I3blocks(s, opts()), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("I3blocks emitted %d lines, want 3: %q", len(lines), lines)
	}
	if lines[0] != "session 90% 1h42 │ week 21% 4d" {
		t.Errorf("full_text = %q", lines[0])
	}
	if lines[1] != "90%│21%" {
		t.Errorf("short_text = %q", lines[1])
	}
	if lines[2] != DefaultColors().Crit {
		t.Errorf("color = %q, want %q", lines[2], DefaultColors().Crit)
	}
}

func TestI3blocksOmitsTheColourLineWhenThereIsNoColour(t *testing.T) {
	calm := live()
	calm.FiveHour.UsedPercentage = 12

	lines := strings.Split(strings.TrimRight(I3blocks(calm, opts()), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("I3blocks emitted %d lines, want 2 (no colour line): %q", len(lines), lines)
	}
}

func TestI3blocksEmitsNothingWhenThereIsNothingToShow(t *testing.T) {
	if got := I3blocks(usage.Snapshot{}, opts()); got != "" {
		t.Errorf("I3blocks() = %q, want empty", got)
	}
}

func TestWaybarCarriesTheDetailInTheTooltip(t *testing.T) {
	var out struct {
		Text       string  `json:"text"`
		Tooltip    string  `json:"tooltip"`
		Class      string  `json:"class"`
		Percentage float64 `json:"percentage"`
	}
	if err := json.Unmarshal([]byte(Waybar(live(), opts())), &out); err != nil {
		t.Fatalf("Waybar output is not valid JSON: %v", err)
	}
	if out.Text != "session 63% 1h42 │ week 21% 4d" {
		t.Errorf("text = %q", out.Text)
	}
	if !strings.Contains(out.Tooltip, "Weekly") {
		t.Errorf("tooltip = %q, want the breakdown", out.Tooltip)
	}
	if out.Class != "warn" {
		t.Errorf("class = %q, want %q", out.Class, "warn")
	}
	if out.Percentage != 63 {
		t.Errorf("percentage = %v, want 63", out.Percentage)
	}
}

func TestPolybarWrapsColourAndClickAction(t *testing.T) {
	s := live()
	s.FiveHour.UsedPercentage = 90

	got := Polybar(s, opts())
	if !strings.Contains(got, "%{F"+DefaultColors().Crit+"}") {
		t.Errorf("Polybar() = %q, want a colour tag", got)
	}
	if !strings.Contains(got, "%{A1:"+DetailCommand+":}") || !strings.HasSuffix(got, "%{A}") {
		t.Errorf("Polybar() = %q, want a left-click action", got)
	}
}

func TestPlainHasNoMarkup(t *testing.T) {
	got := Plain(live(), opts())
	if strings.Contains(got, "%{") || strings.Contains(got, "#") {
		t.Errorf("Plain() = %q, want no markup", got)
	}
}

func TestTemplateReplacesEveryPlaceholder(t *testing.T) {
	o := opts()
	o.Template = "{session_pct}|{session_reset}|{weekly_pct}|{weekly_reset}|{model}|{cost}"

	got := Compact(live(), o).FullText
	want := "63%|1h42|21%|4d|Opus|$2.41"
	if got != want {
		t.Errorf("Compact() with template = %q, want %q", got, want)
	}
}

func TestTemplatePlaceholdersFallBackToTheDashWhenExpired(t *testing.T) {
	s := live()
	s.FiveHour = &usage.Limit{UsedPercentage: 99, ResetsAt: at(-time.Minute)}

	o := opts()
	o.Template = "{session_pct} {session_reset}"
	if got, want := Compact(s, o).FullText, "— —"; got != want {
		t.Errorf("Compact() = %q, want %q", got, want)
	}
}

func TestDetailGolden(t *testing.T) {
	got := Detail(live(), opts())
	golden := filepath.Join("testdata", "detail.txt")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Errorf("Detail() mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestDetailSaysHowOldTheDataIs(t *testing.T) {
	if !strings.Contains(Detail(live(), opts()), "11:58") {
		t.Errorf("Detail() should state the observation time, got:\n%s", Detail(live(), opts()))
	}
}

func TestDetailLinePadsByRuneNotByte(t *testing.T) {
	l := &usage.Limit{UsedPercentage: 63, ResetsAt: at(time.Hour)}

	ascii := detailLine("Weekly", l, now)
	multibyte := detailLine("Sessão", l, now)

	col := func(line string) int { return len([]rune(line[:strings.Index(line, "%")])) }
	if col(ascii) != col(multibyte) {
		t.Errorf("percentage column differs: ascii=%d multibyte=%d — padding must count runes, not bytes\n%s%s",
			col(ascii), col(multibyte), ascii, multibyte)
	}
}

func TestDetailCarriesOnlyAccountWideFacts(t *testing.T) {
	s := live()

	got := Detail(s, opts())
	for _, leaked := range []string{"Opus", "2.41", "Context", "45m"} {
		if strings.Contains(got, leaked) {
			t.Errorf("Detail() leaked %q — per-session figures belong to whichever session wrote last, which is arbitrary with several open:\n%s", leaked, got)
		}
	}
	for _, want := range []string{"5h window", "Weekly", "as of"} {
		if !strings.Contains(got, want) {
			t.Errorf("Detail() is missing %q:\n%s", want, got)
		}
	}
}

func TestJSONStillExposesTheSessionForCustomFormats(t *testing.T) {
	var out struct {
		Session *struct {
			Model   string  `json:"model"`
			CostUSD float64 `json:"cost_usd"`
		} `json:"session"`
	}
	if err := json.Unmarshal([]byte(JSON(live(), opts())), &out); err != nil {
		t.Fatalf("JSON output is not valid: %v", err)
	}
	if out.Session == nil || out.Session.Model != "Opus" {
		t.Errorf("session = %+v, want it still available to --format json and --template", out.Session)
	}
}
