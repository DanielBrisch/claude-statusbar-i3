package render

import (
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

type windowView struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       string  `json:"resets_at"`
	TimeLeft       string  `json:"time_left"`
	Expired        bool    `json:"expired"`
}

func newWindowView(l *usage.Limit, now time.Time) *windowView {
	if l == nil {
		return nil
	}
	return &windowView{
		UsedPercentage: l.UsedPercentage,
		ResetsAt:       newStamp(l.ResetsAt).String(),
		TimeLeft:       l.TimeLeft(now).String(),
		Expired:        l.Expired(now),
	}
}

type sessionView struct {
	Model             string  `json:"model"`
	CostUSD           float64 `json:"cost_usd"`
	ContextUsedPct    float64 `json:"context_used_pct"`
	ContextWindowSize int64   `json:"context_window_size"`
	DurationMinutes   int     `json:"duration_minutes"`
}

func newSessionView(s *usage.Session) *sessionView {
	if s == nil {
		return nil
	}
	return &sessionView{
		Model:             s.Model,
		CostUSD:           s.CostUSD,
		ContextUsedPct:    s.ContextUsedPct,
		ContextWindowSize: s.ContextWindowSize,
		DurationMinutes:   int(s.Duration / time.Minute),
	}
}

type snapshotView struct {
	Text       string       `json:"text"`
	Level      string       `json:"level"`
	Percentage float64      `json:"percentage"`
	FiveHour   *windowView  `json:"five_hour"`
	SevenDay   *windowView  `json:"seven_day"`
	SpendLimit *windowView  `json:"spend_limit"`
	Session    *sessionView `json:"session"`
	UpdatedAt  string       `json:"updated_at"`
}
