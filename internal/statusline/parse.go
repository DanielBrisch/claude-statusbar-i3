package statusline

import (
	"encoding/json"
	"fmt"
	"io"
)

type Limit struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

type RateLimits struct {
	FiveHour   *Limit `json:"five_hour"`
	SevenDay   *Limit `json:"seven_day"`
	SpendLimit *Limit `json:"spend_limit"`
}

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

func Parse(r io.Reader) (Payload, error) {
	var w wirePayload
	if err := json.NewDecoder(r).Decode(&w); err != nil {
		return Payload{}, fmt.Errorf("decode statusline payload: %w", err)
	}
	return Payload{
		SessionID:         w.SessionID,
		Model:             w.Model.DisplayName,
		CWD:               w.CWD,
		CostUSD:           w.Cost.TotalCostUSD,
		DurationMS:        w.Cost.TotalDurationMS,
		ContextUsedPct:    w.ContextWindow.UsedPercentage,
		ContextWindowSize: w.ContextWindow.ContextWindowSize,
		RateLimits:        w.RateLimits,
	}, nil
}
