package usage

type Thresholds struct {
	Warn   float64
	Crit   float64
	Urgent float64
}

func NewThresholds(warn, crit, urgent float64) Thresholds {
	return Thresholds{Warn: warn, Crit: crit, Urgent: urgent}
}

func DefaultThresholds() Thresholds {
	return NewThresholds(60, 85, 95)
}

func (t Thresholds) Level(pct float64) Level {
	switch {
	case pct >= t.Urgent:
		return LevelUrgent
	case pct >= t.Crit:
		return LevelCrit
	case pct >= t.Warn:
		return LevelWarn
	default:
		return LevelOK
	}
}
