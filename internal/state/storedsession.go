package state

import (
	"time"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/statusline"
	"github.com/DanielBrisch/claude-statusbar-i3/internal/usage"
)

type storedSession struct {
	Model             string  `json:"model"`
	CostUSD           float64 `json:"cost_usd"`
	ContextUsedPct    float64 `json:"context_used_pct"`
	ContextWindowSize int64   `json:"context_window_size"`
	DurationMS        int64   `json:"duration_ms"`
	CWD               string  `json:"cwd"`
	UpdatedAt         int64   `json:"updated_at"`
	Seq               int64   `json:"seq"`
}

func newStoredSession(p statusline.Payload, at time.Time, seq int64) storedSession {
	return storedSession{
		Model:             p.Model,
		CostUSD:           p.CostUSD,
		ContextUsedPct:    p.ContextUsedPct,
		ContextWindowSize: p.ContextWindowSize,
		DurationMS:        p.DurationMS,
		CWD:               p.CWD,
		UpdatedAt:         at.Unix(),
		Seq:               seq,
	}
}

func (s storedSession) idleFor(now time.Time) time.Duration {
	return now.Sub(time.Unix(s.UpdatedAt, 0))
}

func (s storedSession) domain() *usage.Session {
	return usage.NewSession(
		s.Model, s.CostUSD, s.ContextUsedPct, s.ContextWindowSize,
		time.Duration(s.DurationMS)*time.Millisecond, s.CWD, time.Unix(s.UpdatedAt, 0),
	)
}
