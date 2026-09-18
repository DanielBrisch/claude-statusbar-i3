package usage

type Level int

const (
	LevelOK Level = iota
	LevelWarn
	LevelCrit
	LevelUrgent
)

func (l Level) String() string {
	switch l {
	case LevelWarn:
		return "warn"
	case LevelCrit:
		return "crit"
	case LevelUrgent:
		return "urgent"
	default:
		return "ok"
	}
}

func (l Level) WorseThan(other Level) bool {
	return l > other
}
