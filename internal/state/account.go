package state

import (
	"time"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/statusline"
)

type account struct {
	FiveHour   *storedLimit `json:"five_hour"`
	SevenDay   *storedLimit `json:"seven_day"`
	SpendLimit *storedLimit `json:"spend_limit"`
}

func newAccount(r *statusline.RateLimits, observedAt time.Time) account {
	if r == nil {
		return account{}
	}
	return account{
		FiveHour:   newStoredLimit(r.FiveHour, observedAt),
		SevenDay:   newStoredLimit(r.SevenDay, observedAt),
		SpendLimit: newStoredLimit(r.SpendLimit, observedAt),
	}
}
