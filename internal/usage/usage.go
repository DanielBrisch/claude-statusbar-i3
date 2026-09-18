package usage

import (
	"fmt"
	"time"
)

type Level int

const (
	LevelOK Level = iota
	LevelWarn
	LevelCrit
	LevelUrgent
)

func (l Level) String() string {
	switch l {
	case LevelWarn:
		return "warn"
	case LevelCrit:
		return "crit"
	case LevelUrgent:
		return "urgent"
	default:
		return "ok"
	}
}

type Thresholds struct {
	Warn   float64
	Crit   float64
	Urgent float64
}

func DefaultThresholds() Thresholds {
	return Thresholds{Warn: 60, Crit: 85, Urgent: 95}
}

func (t Thresholds) Level(pct float64) Level {
	switch {
	case pct >= t.Urgent:
		return LevelUrgent
	case pct >= t.Crit:
		return LevelCrit
	case pct >= t.Warn:
		return LevelWarn
	default:
		return LevelOK
	}
}

type Limit struct {
	UsedPercentage float64
	ResetsAt       time.Time
	ObservedAt     time.Time
}

func (l Limit) Expired(now time.Time) bool {
	return !now.Before(l.ResetsAt)
}

func (l Limit) TimeLeft(now time.Time) time.Duration {
	if l.Expired(now) {
		return 0
	}
	return l.ResetsAt.Sub(now)
}

type Session struct {
	Model             string
	CostUSD           float64
	ContextUsedPct    float64
	ContextWindowSize int64
	Duration          time.Duration
	CWD               string
	UpdatedAt         time.Time
}

type Snapshot struct {
	FiveHour   *Limit
	SevenDay   *Limit
	SpendLimit *Limit
	Session    *Session
	UpdatedAt  time.Time
}

func (s Snapshot) visible(now time.Time) []*Limit {
	var out []*Limit
	for _, l := range []*Limit{s.FiveHour, s.SevenDay, s.SpendLimit} {
		if l != nil && !l.Expired(now) {
			out = append(out, l)
		}
	}
	return out
}

func (s Snapshot) HasVisibleWindow(now time.Time) bool {
	return len(s.visible(now)) > 0
}

func (s Snapshot) Level(now time.Time, t Thresholds) Level {
	worst := LevelOK
	for _, l := range s.visible(now) {
		if lv := t.Level(l.UsedPercentage); lv > worst {
			worst = lv
		}
	}
	return worst
}

func FormatTimeLeft(d time.Duration) string {
	if d >= 24*time.Hour {
		return fmt.Sprintf("%dd", int(d/(24*time.Hour)))
	}
	if d >= time.Hour {
		return fmt.Sprintf("%dh%02d", int(d/time.Hour), int(d%time.Hour/time.Minute))
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm", int(d/time.Minute))
	}
	return "<1m"
}
