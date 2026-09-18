package usage

import "time"

type Snapshot struct {
	FiveHour   *Limit
	SevenDay   *Limit
	SpendLimit *Limit
	Session    *Session
	UpdatedAt  time.Time
}

func NewSnapshot(fiveHour, sevenDay, spendLimit *Limit, session *Session, updatedAt time.Time) Snapshot {
	return Snapshot{
		FiveHour:   fiveHour,
		SevenDay:   sevenDay,
		SpendLimit: spendLimit,
		Session:    session,
		UpdatedAt:  updatedAt,
	}
}

func (s Snapshot) Windows() []*Limit {
	return []*Limit{s.FiveHour, s.SevenDay, s.SpendLimit}
}

func (s Snapshot) Live(now time.Time) []*Limit {
	var out []*Limit
	for _, l := range s.Windows() {
		if l.Live(now) {
			out = append(out, l)
		}
	}
	return out
}

func (s Snapshot) HasVisibleWindow(now time.Time) bool {
	return len(s.Live(now)) > 0
}

func (s Snapshot) Level(now time.Time, t Thresholds) Level {
	worst := LevelOK
	for _, l := range s.Live(now) {
		if lv := t.Level(l.UsedPercentage); lv.WorseThan(worst) {
			worst = lv
		}
	}
	return worst
}

func (s Snapshot) PrimaryPercentage(now time.Time) float64 {
	for _, l := range []*Limit{s.FiveHour, s.SevenDay} {
		if l.Live(now) {
			return l.UsedPercentage
		}
	}
	return 0
}
