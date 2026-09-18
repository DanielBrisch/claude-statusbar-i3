package usage

import (
	"testing"
	"time"
)

var now = time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)

func at(d time.Duration) time.Time { return now.Add(d) }

func TestLimitExpiredWhenResetHasPassed(t *testing.T) {
	cases := []struct {
		name     string
		resetsAt time.Time
		want     bool
	}{
		{"reset in the future", at(90 * time.Minute), false},
		{"reset exactly now", now, true},
		{"reset in the past", at(-time.Second), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := Limit{ResetsAt: tc.resetsAt}
			if got := l.Expired(now); got != tc.want {
				t.Errorf("Expired() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLimitTimeLeftIsZeroOnceExpired(t *testing.T) {
	l := Limit{ResetsAt: at(-time.Hour)}
	if got := l.TimeLeft(now); got != 0 {
		t.Errorf("TimeLeft() = %v, want 0", got)
	}
}

func TestTimeLeftString(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{102 * time.Minute, "1h42"},
		{5 * time.Hour, "5h00"},
		{42 * time.Minute, "42m"},
		{30 * time.Second, "<1m"},
		{0, "<1m"},
		{4 * 24 * time.Hour, "4d"},
		{4*24*time.Hour + 13*time.Hour, "4d"},
		{23*time.Hour + 59*time.Minute, "23h59"},
	}
	for _, tc := range cases {
		if got := NewTimeLeft(tc.in).String(); got != tc.want {
			t.Errorf("NewTimeLeft(%v).String() = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestLevelFromThresholds(t *testing.T) {
	th := Thresholds{Warn: 60, Crit: 85, Urgent: 95}
	cases := []struct {
		pct  float64
		want Level
	}{
		{0, LevelOK},
		{59.9, LevelOK},
		{60, LevelWarn},
		{84.9, LevelWarn},
		{85, LevelCrit},
		{94.9, LevelCrit},
		{95, LevelUrgent},
		{140, LevelUrgent},
	}
	for _, tc := range cases {
		if got := th.Level(tc.pct); got != tc.want {
			t.Errorf("Level(%v) = %v, want %v", tc.pct, got, tc.want)
		}
	}
}

func TestSnapshotLevelUsesTheWorstVisibleWindow(t *testing.T) {
	th := Thresholds{Warn: 60, Crit: 85, Urgent: 95}
	s := Snapshot{
		FiveHour: &Limit{UsedPercentage: 10, ResetsAt: at(time.Hour)},
		SevenDay: &Limit{UsedPercentage: 88, ResetsAt: at(72 * time.Hour)},
	}
	if got := s.Level(now, th); got != LevelCrit {
		t.Errorf("Level() = %v, want %v", got, LevelCrit)
	}
}

func TestSnapshotLevelIgnoresExpiredWindows(t *testing.T) {
	th := Thresholds{Warn: 60, Crit: 85, Urgent: 95}
	s := Snapshot{
		FiveHour: &Limit{UsedPercentage: 99, ResetsAt: at(-time.Minute)},
		SevenDay: &Limit{UsedPercentage: 10, ResetsAt: at(72 * time.Hour)},
	}
	if got := s.Level(now, th); got != LevelOK {
		t.Errorf("Level() = %v, want %v (expired 5h window must not colour the bar)", got, LevelOK)
	}
}

func TestSnapshotHasNothingToShowWhenEveryWindowIsAbsentOrExpired(t *testing.T) {
	empty := Snapshot{}
	if empty.HasVisibleWindow(now) {
		t.Error("HasVisibleWindow() = true for an empty snapshot, want false")
	}

	expired := Snapshot{
		FiveHour: &Limit{ResetsAt: at(-time.Hour)},
		SevenDay: &Limit{ResetsAt: at(-time.Hour)},
	}
	if expired.HasVisibleWindow(now) {
		t.Error("HasVisibleWindow() = true when every window expired, want false")
	}

	weeklyOnly := Snapshot{
		FiveHour: &Limit{ResetsAt: at(-time.Hour)},
		SevenDay: &Limit{UsedPercentage: 21, ResetsAt: at(96 * time.Hour)},
	}
	if !weeklyOnly.HasVisibleWindow(now) {
		t.Error("HasVisibleWindow() = false while the weekly window is still live, want true")
	}
}
