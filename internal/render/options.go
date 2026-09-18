package render

import (
	"fmt"
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

const (
	Placeholder        = "—"
	DefaultLabel       = "✳"
	DefaultWeeklyLabel = "week"
	DefaultIconSize    = "large"
	DetailCommand      = "claude-statusbar detail --notify"
)

type Options struct {
	Now         time.Time
	Label       string
	WeeklyLabel string
	Template    string
	Markup      Markup
	IconSize    string
	UrgentExit  bool
	Thresholds  usage.Thresholds
	Colors      Colors
}

func NewOptions(now time.Time) Options {
	return Options{
		Now:         now,
		Label:       DefaultLabel,
		WeeklyLabel: DefaultWeeklyLabel,
		Markup:      MarkupNone,
		IconSize:    DefaultIconSize,
		Thresholds:  usage.DefaultThresholds(),
		Colors:      DefaultColors(),
	}
}

func (o Options) Icon() string {
	if !o.Markup.Pango() || o.Label == "" {
		return o.Label
	}
	size := o.IconSize
	if size == "" {
		size = DefaultIconSize
	}
	return fmt.Sprintf("<span size=%q>%s</span>", size, o.Markup.Escape(o.Label))
}

func (o Options) Escape(s string) string { return o.Markup.Escape(s) }

func (o Options) Plain() Options {
	o.Markup = MarkupNone
	return o
}
