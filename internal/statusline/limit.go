package statusline

type Limit struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

func NewLimit(usedPercentage float64, resetsAt int64) *Limit {
	return &Limit{UsedPercentage: usedPercentage, ResetsAt: resetsAt}
}
