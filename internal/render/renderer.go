package render

import (
	"strings"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/usage"
)

type Renderer struct {
	opts Options
}

func NewRenderer(o Options) *Renderer {
	return &Renderer{opts: o}
}

func (r *Renderer) Options() Options { return r.opts }

func (r *Renderer) Block(s usage.Snapshot) Block {
	if !s.HasVisibleWindow(r.opts.Now) {
		return Block{}
	}

	level := s.Level(r.opts.Now, r.opts.Thresholds)
	five := newSegment(r.opts.Icon(), s.FiveHour, r.opts)
	seven := newSegment(r.opts.Escape(r.opts.WeeklyLabel), s.SevenDay, r.opts)

	b := Block{
		Level:      level,
		Color:      r.opts.Colors.For(level),
		Percentage: s.PrimaryPercentage(r.opts.Now),
		ShortText:  five.percent() + "│" + seven.percent(),
	}

	if tmpl := NewTemplate(r.opts.Template, r.opts); !tmpl.Empty() {
		b.FullText = tmpl.Expand(s)
		return b
	}
	b.FullText = strings.TrimSpace(strings.Join(
		[]string{five.String(), r.opts.Escape("│"), seven.String()}, " "))
	return b
}

func (r *Renderer) Detail(s usage.Snapshot) Detail {
	return NewDetail(s, r.opts)
}
