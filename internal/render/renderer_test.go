package render

import (
	"strings"
	"testing"
	"time"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/usage"
)

func TestBlockShowsBothWindows(t *testing.T) {
	if got, want := block(live(), opts()).FullText, "✳ session 63% 1h42 │ week 21% 4d"; got != want {
		t.Errorf("FullText = %q, want %q", got, want)
	}
}

func TestBlockPlaceholdersTheSessionButKeepsTheWeekly(t *testing.T) {
	s := live()
	s.FiveHour = usage.NewLimit(99, at(-time.Minute), now)

	if got, want := block(s, opts()).FullText, "✳ session — │ week 21% 4d"; got != want {
		t.Errorf("FullText = %q, want %q", got, want)
	}
}

func TestBlockIsEmptyWhenNothingIsLive(t *testing.T) {
	s := usage.NewSnapshot(
		usage.NewLimit(0, at(-time.Hour), now),
		usage.NewLimit(0, at(-time.Hour), now),
		nil, nil, now,
	)
	if got := block(s, opts()).FullText; got != "" {
		t.Errorf("FullText = %q, want empty", got)
	}
}

func TestBlockWithoutAnyRateLimitsIsEmpty(t *testing.T) {
	if got := block(usage.Snapshot{}, opts()).FullText; got != "" {
		t.Errorf("FullText = %q, want empty", got)
	}
}

func TestBlockStaysUncolouredByDefault(t *testing.T) {
	for _, pct := range []float64{10, 70, 90, 97, 140} {
		s := live()
		s.FiveHour.UsedPercentage = pct
		if got := block(s, opts()).Color; got != "" {
			t.Errorf("at %v%% Color = %q, want empty — the bar's own statusline colour is the default", pct, got)
		}
	}
}

func TestBlockColourFollowsTheWorstWindowWhenColoursAreConfigured(t *testing.T) {
	cases := []struct {
		name      string
		fiveHour  float64
		wantColor string
		wantLevel usage.Level
	}{
		{"calm", 10, "", usage.LevelOK},
		{"warn", 70, "#E5C07B", usage.LevelWarn},
		{"crit", 90, "#E06C75", usage.LevelCrit},
		{"urgent", 97, "#E06C75", usage.LevelUrgent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := live()
			s.FiveHour.UsedPercentage = tc.fiveHour
			b := block(s, coloured(opts()))
			if b.Color != tc.wantColor {
				t.Errorf("Color = %q, want %q", b.Color, tc.wantColor)
			}
			if b.Level != tc.wantLevel {
				t.Errorf("Level = %v, want %v", b.Level, tc.wantLevel)
			}
		})
	}
}

func TestTemplateReplacesEveryPlaceholder(t *testing.T) {
	o := opts()
	o.Template = "{session_pct}|{session_reset}|{weekly_pct}|{weekly_reset}|{model}|{cost}"

	if got, want := block(live(), o).FullText, "63%|1h42|21%|4d|Opus|$2.41"; got != want {
		t.Errorf("FullText = %q, want %q", got, want)
	}
}

func TestTemplatePlaceholdersFallBackToTheDashWhenExpired(t *testing.T) {
	s := live()
	s.FiveHour = usage.NewLimit(99, at(-time.Minute), now)

	o := opts()
	o.Template = "{session_pct} {session_reset}"
	if got, want := block(s, o).FullText, "— —"; got != want {
		t.Errorf("FullText = %q, want %q", got, want)
	}
}

func TestPangoMarkupEnlargesOnlyTheIcon(t *testing.T) {
	o := opts()
	o.Markup = MarkupPango
	o.IconSize = "x-large"

	want := `<span size="x-large">✳</span> session 63% 1h42 │ week 21% 4d`
	if got := block(live(), o).FullText; got != want {
		t.Errorf("FullText = %q\nwant %q", got, want)
	}
}

func TestPangoMarkupEscapesEverythingElse(t *testing.T) {
	o := opts()
	o.Markup = MarkupPango
	o.Label = "a&b"
	o.WeeklyLabel = "<w>"

	got := block(live(), o).FullText
	if strings.Contains(got, "a&b") || strings.Contains(got, "<w>") {
		t.Errorf("FullText = %q — pango markup must escape the text, or a stray & breaks i3bar's parser", got)
	}
	if !strings.Contains(got, "a&amp;b") || !strings.Contains(got, "&lt;w&gt;") {
		t.Errorf("FullText = %q, want escaped label and weekly label", got)
	}
}

func TestWithoutPangoTheTextStaysLiteral(t *testing.T) {
	o := opts()
	o.IconSize = "x-large"

	got := block(live(), o).FullText
	if strings.Contains(got, "<span") {
		t.Errorf("FullText = %q — without markup=pango a span tag would show up literally on the bar", got)
	}
	if got != "✳ session 63% 1h42 │ week 21% 4d" {
		t.Errorf("FullText = %q", got)
	}
}

func TestDefaultLabelCarriesNoEmojiVariationSelector(t *testing.T) {
	for _, r := range DefaultLabel {
		if r == '️' {
			t.Fatalf("DefaultLabel %q carries U+FE0F, which forces emoji presentation: the glyph would come from the colour emoji font, ignore the block colour and break the monospace width", DefaultIcon)
		}
	}
}

func TestTheSessionWindowIsNamedJustLikeTheWeeklyOne(t *testing.T) {
	got := block(live(), opts()).FullText

	if !strings.Contains(got, "session 63%") {
		t.Errorf("FullText = %q — the icon says which tool, the word says which window, the same way week does", got)
	}
	if strings.Index(got, "session") >= strings.Index(got, "week") {
		t.Errorf("FullText = %q, want the session segment first", got)
	}
}

func TestTheIconAndTheWordAreSeparateKnobs(t *testing.T) {
	o := opts()
	o.Icon = ""
	if got, want := block(live(), o).FullText, "session 63% 1h42 │ week 21% 4d"; got != want {
		t.Errorf("without an icon FullText = %q, want %q", got, want)
	}

	o = opts()
	o.Label = ""
	if got, want := block(live(), o).FullText, "✳ 63% 1h42 │ week 21% 4d"; got != want {
		t.Errorf("without a label FullText = %q, want %q", got, want)
	}
}

func TestOnlyTheIconGrowsNotTheWord(t *testing.T) {
	o := opts()
	o.Markup = MarkupPango
	o.IconSize = "x-large"

	got := block(live(), o).FullText
	if !strings.Contains(got, `<span size="x-large">✳</span> session`) {
		t.Errorf("FullText = %q — the span must close before the word, or 'session' is enlarged too", got)
	}
	if strings.Count(got, "<span") != 1 {
		t.Errorf("FullText = %q, want exactly one span", got)
	}
}

func TestTemplateSeparatesIconFromLabel(t *testing.T) {
	o := opts()
	o.Template = "{icon}/{label}/{weekly_label}"

	if got, want := block(live(), o).FullText, "✳/session/week"; got != want {
		t.Errorf("FullText = %q, want %q", got, want)
	}
}
