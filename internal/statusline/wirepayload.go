package statusline

type wirePayload struct {
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
	Model     struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Cost struct {
		TotalCostUSD    float64 `json:"total_cost_usd"`
		TotalDurationMS int64   `json:"total_duration_ms"`
	} `json:"cost"`
	ContextWindow struct {
		UsedPercentage    float64 `json:"used_percentage"`
		ContextWindowSize int64   `json:"context_window_size"`
	} `json:"context_window"`
	RateLimits *RateLimits `json:"rate_limits"`
}

func (w wirePayload) payload() Payload {
	return Payload{
		SessionID:         w.SessionID,
		Model:             w.Model.DisplayName,
		CWD:               w.CWD,
		CostUSD:           w.Cost.TotalCostUSD,
		DurationMS:        w.Cost.TotalDurationMS,
		ContextUsedPct:    w.ContextWindow.UsedPercentage,
		ContextWindowSize: w.ContextWindow.ContextWindowSize,
		RateLimits:        w.RateLimits,
	}
}
