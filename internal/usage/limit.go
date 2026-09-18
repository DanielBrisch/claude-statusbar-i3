package usage

import (
	"fmt"
	"time"
)

type Limit struct {
	UsedPercentage float64
	ResetsAt       time.Time
	ObservedAt     time.Time
}

func NewLimit(usedPercentage float64, resetsAt, observedAt time.Time) *Limit {
	return &Limit{UsedPercentage: usedPercentage, ResetsAt: resetsAt, ObservedAt: observedAt}
}

func (l *Limit) Expired(now time.Time) bool {
	return l == nil || !now.Before(l.ResetsAt)
}

func (l *Limit) Live(now time.Time) bool {
	return !l.Expired(now)
}

func (l *Limit) TimeLeft(now time.Time) TimeLeft {
	if l.Expired(now) {
		return 0
	}
	return NewTimeLeft(l.ResetsAt.Sub(now))
}

func (l *Limit) Percent() string {
	if l == nil {
		return ""
	}
	return fmt.Sprintf("%.0f%%", l.UsedPercentage)
}

func (l *Limit) Level(now time.Time, t Thresholds) Level {
	if l.Expired(now) {
		return LevelOK
	}
	return t.Level(l.UsedPercentage)
}
