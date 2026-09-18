package state

import (
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/statusline"
	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

type storedLimit struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
	ObservedAt     int64   `json:"observed_at"`
}

func newStoredLimit(l *statusline.Limit, observedAt time.Time) *storedLimit {
	if l == nil {
		return nil
	}
	return &storedLimit{
		UsedPercentage: l.UsedPercentage,
		ResetsAt:       l.ResetsAt,
		ObservedAt:     observedAt.Unix(),
	}
}

func (l *storedLimit) domain() *usage.Limit {
	if l == nil {
		return nil
	}
	return usage.NewLimit(l.UsedPercentage, time.Unix(l.ResetsAt, 0), time.Unix(l.ObservedAt, 0))
}
