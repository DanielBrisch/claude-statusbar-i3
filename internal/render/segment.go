package render

import (
	"github.com/DanielBrisch/claude-statusbar-i3/internal/usage"
)

type segment struct {
	label string
	limit *usage.Limit
	opts  Options
}

func newSegment(label string, l *usage.Limit, o Options) segment {
	return segment{label: label, limit: l, opts: o}
}

func (s segment) percent() string {
	if !s.limit.Known() {
		return Placeholder
	}
	return s.limit.PercentAt(s.opts.Now)
}

func (s segment) reset() string {
	if !s.limit.Live(s.opts.Now) {
		return Placeholder
	}
	return s.limit.TimeLeft(s.opts.Now).String()
}

func (s segment) String() string {
	body := Placeholder
	switch {
	case s.limit.Live(s.opts.Now):
		body = s.percent() + " " + s.reset()
	case s.limit.Known():
		body = s.percent()
	}
	body = s.opts.Escape(body)
	if s.label == "" {
		return body
	}
	return s.label + " " + body
}
