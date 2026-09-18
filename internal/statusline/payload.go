package statusline

type Payload struct {
	SessionID         string
	Model             string
	CWD               string
	CostUSD           float64
	DurationMS        int64
	ContextUsedPct    float64
	ContextWindowSize int64
	RateLimits        *RateLimits
}

func (p Payload) HasSession() bool {
	return p.SessionID != ""
}

func (p Payload) HasRateLimits() bool {
	return p.RateLimits != nil
}
