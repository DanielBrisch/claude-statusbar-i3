package usage

import "time"

type Session struct {
	Model             string
	CostUSD           float64
	ContextUsedPct    float64
	ContextWindowSize int64
	Duration          time.Duration
	CWD               string
	UpdatedAt         time.Time
}

func NewSession(model string, costUSD, contextUsedPct float64, contextWindowSize int64, duration time.Duration, cwd string, updatedAt time.Time) *Session {
	return &Session{
		Model:             model,
		CostUSD:           costUSD,
		ContextUsedPct:    contextUsedPct,
		ContextWindowSize: contextWindowSize,
		Duration:          duration,
		CWD:               cwd,
		UpdatedAt:         updatedAt,
	}
}

func (s *Session) ContextWindowThousands() int64 {
	if s == nil {
		return 0
	}
	return s.ContextWindowSize / 1000
}
