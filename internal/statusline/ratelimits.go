package statusline

type RateLimits struct {
	FiveHour   *Limit `json:"five_hour"`
	SevenDay   *Limit `json:"seven_day"`
	SpendLimit *Limit `json:"spend_limit"`
}

func NewRateLimits(fiveHour, sevenDay, spendLimit *Limit) *RateLimits {
	return &RateLimits{FiveHour: fiveHour, SevenDay: sevenDay, SpendLimit: spendLimit}
}

func (r *RateLimits) Empty() bool {
	return r == nil || (r.FiveHour == nil && r.SevenDay == nil && r.SpendLimit == nil)
}
