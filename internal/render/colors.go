package render

import "github.com/DanielBrisch/claude-statusbar-i3/internal/usage"

type Colors struct {
	OK   string
	Warn string
	Crit string
}

func NewColors(warn, crit string) Colors {
	return Colors{Warn: warn, Crit: crit}
}

func DefaultColors() Colors {
	return Colors{}
}

func (c Colors) For(l usage.Level) string {
	switch l {
	case usage.LevelWarn:
		return c.Warn
	case usage.LevelCrit, usage.LevelUrgent:
		return c.Crit
	default:
		return c.OK
	}
}
