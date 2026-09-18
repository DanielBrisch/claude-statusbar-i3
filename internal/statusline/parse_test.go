package statusline

import (
	"os"
	"strings"
	"testing"
)

func parseFixture(t *testing.T, name string) Payload {
	t.Helper()
	f, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()
	p, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse(%s): %v", name, err)
	}
	return p
}

func TestParseReadsRateLimits(t *testing.T) {
	p := parseFixture(t, "full.json")

	if p.RateLimits == nil {
		t.Fatal("RateLimits is nil, want populated")
	}
	if p.RateLimits.FiveHour == nil {
		t.Fatal("FiveHour is nil, want populated")
	}
	if got, want := p.RateLimits.FiveHour.UsedPercentage, 23.5; got != want {
		t.Errorf("FiveHour.UsedPercentage = %v, want %v", got, want)
	}
	if got, want := p.RateLimits.FiveHour.ResetsAt, int64(1738425600); got != want {
		t.Errorf("FiveHour.ResetsAt = %v, want %v", got, want)
	}
	if got, want := p.RateLimits.SevenDay.UsedPercentage, 41.2; got != want {
		t.Errorf("SevenDay.UsedPercentage = %v, want %v", got, want)
	}
	if got, want := p.RateLimits.SpendLimit.ResetsAt, int64(1740787200); got != want {
		t.Errorf("SpendLimit.ResetsAt = %v, want %v", got, want)
	}
}

func TestParseReadsSessionFields(t *testing.T) {
	p := parseFixture(t, "full.json")

	if got, want := p.SessionID, "abc123"; got != want {
		t.Errorf("SessionID = %q, want %q", got, want)
	}
	if got, want := p.Model, "Opus"; got != want {
		t.Errorf("Model = %q, want %q", got, want)
	}
	if got, want := p.CostUSD, 1.23; got != want {
		t.Errorf("CostUSD = %v, want %v", got, want)
	}
	if got, want := p.DurationMS, int64(45000); got != want {
		t.Errorf("DurationMS = %v, want %v", got, want)
	}
	if got, want := p.ContextUsedPct, 8.0; got != want {
		t.Errorf("ContextUsedPct = %v, want %v", got, want)
	}
	if got, want := p.ContextWindowSize, int64(200000); got != want {
		t.Errorf("ContextWindowSize = %v, want %v", got, want)
	}
	if got, want := p.CWD, "/home/daniel/proj"; got != want {
		t.Errorf("CWD = %q, want %q", got, want)
	}
}

func TestParseWithoutRateLimitsIsNotAnError(t *testing.T) {
	p := parseFixture(t, "no_rate_limits.json")

	if p.RateLimits != nil {
		t.Errorf("RateLimits = %+v, want nil for a payload without rate_limits", p.RateLimits)
	}
	if got, want := p.SessionID, "def456"; got != want {
		t.Errorf("SessionID = %q, want %q", got, want)
	}
}

func TestParseKeepsAbsentWindowsNil(t *testing.T) {
	p := parseFixture(t, "partial_rate_limits.json")

	if p.RateLimits == nil || p.RateLimits.FiveHour == nil {
		t.Fatal("FiveHour should be populated")
	}
	if p.RateLimits.SevenDay != nil {
		t.Errorf("SevenDay = %+v, want nil when absent from the payload", p.RateLimits.SevenDay)
	}
	if p.RateLimits.SpendLimit != nil {
		t.Errorf("SpendLimit = %+v, want nil when absent from the payload", p.RateLimits.SpendLimit)
	}
}

func TestParseZeroPercentageIsPreservedNotDropped(t *testing.T) {
	p := parseFixture(t, "partial_rate_limits.json")

	if got, want := p.RateLimits.FiveHour.UsedPercentage, 0.0; got != want {
		t.Errorf("FiveHour.UsedPercentage = %v, want %v", got, want)
	}
}

func TestParseRejectsMalformedJSON(t *testing.T) {
	if _, err := Parse(strings.NewReader("{not json")); err == nil {
		t.Fatal("Parse of malformed JSON returned nil error, want an error")
	}
}
