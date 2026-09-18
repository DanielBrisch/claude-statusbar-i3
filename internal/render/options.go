package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/usage"
)

const (
	Placeholder        = "—"
	DefaultIcon        = "✳"
	DefaultLabel       = "session"
	DefaultWeeklyLabel = "week"
	DefaultIconSize    = "large"
	DetailCommand      = "claude-statusbar detail --notify"
)

type Options struct {
	Now         time.Time
	Icon        string
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
		Icon:        DefaultIcon,
		Label:       DefaultLabel,
		WeeklyLabel: DefaultWeeklyLabel,
		Markup:      MarkupNone,
		IconSize:    DefaultIconSize,
		Thresholds:  usage.DefaultThresholds(),
		Colors:      DefaultColors(),
	}
}

func (o Options) RenderedIcon() string {
	if !o.Markup.Pango() || o.Icon == "" {
		return o.Icon
	}
	size := o.IconSize
	if size == "" {
		size = DefaultIconSize
	}
	return fmt.Sprintf("<span size=%q>%s</span>", size, o.Markup.Escape(o.Icon))
}

func (o Options) SessionPrefix() string {
	parts := make([]string, 0, 2)
	if icon := o.RenderedIcon(); icon != "" {
		parts = append(parts, icon)
	}
	if o.Label != "" {
		parts = append(parts, o.Escape(o.Label))
	}
	return strings.Join(parts, " ")
}

func (o Options) Escape(s string) string { return o.Markup.Escape(s) }

func (o Options) Plain() Options {
	o.Markup = MarkupNone
	return o
}
