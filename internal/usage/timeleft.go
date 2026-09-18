package usage

import (
	"fmt"
	"time"
)

type TimeLeft time.Duration

func NewTimeLeft(d time.Duration) TimeLeft {
	if d < 0 {
		return 0
	}
	return TimeLeft(d)
}

func (t TimeLeft) Duration() time.Duration { return time.Duration(t) }

func (t TimeLeft) String() string {
	d := t.Duration()
	switch {
	case d >= 24*time.Hour:
		return fmt.Sprintf("%dd", int(d/(24*time.Hour)))
	case d >= time.Hour:
		return fmt.Sprintf("%dh%02d", int(d/time.Hour), int(d%time.Hour/time.Minute))
	case d >= time.Minute:
		return fmt.Sprintf("%dm", int(d/time.Minute))
	default:
		return "<1m"
	}
}
