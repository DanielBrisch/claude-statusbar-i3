package render

import "time"

type stamp time.Time

func newStamp(t time.Time) stamp { return stamp(t) }

func (s stamp) known() bool {
	t := time.Time(s)
	return !t.IsZero() && t.Unix() > 0
}

func (s stamp) String() string {
	if !s.known() {
		return ""
	}
	return time.Time(s).Format(time.RFC3339)
}

func (s stamp) Clock() string {
	if !s.known() {
		return "?"
	}
	return time.Time(s).Format("15:04")
}
