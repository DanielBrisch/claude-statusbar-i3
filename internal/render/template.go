package render

import (
	"fmt"
	"strings"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

type Template struct {
	text string
	opts Options
}

func NewTemplate(text string, o Options) Template {
	return Template{text: text, opts: o}
}

func (t Template) Empty() bool { return t.text == "" }

func (t Template) Expand(s usage.Snapshot) string {
	model, cost, context := "", "", ""
	if s.Session != nil {
		model = s.Session.Model
		cost = fmt.Sprintf("$%.2f", s.Session.CostUSD)
		context = fmt.Sprintf("%.0f%%", s.Session.ContextUsedPct)
	}
	five := newSegment("", s.FiveHour, t.opts)
	seven := newSegment("", s.SevenDay, t.opts)
	spend := newSegment("", s.SpendLimit, t.opts)

	return strings.NewReplacer(
		"{label}", t.opts.Icon(),
		"{weekly_label}", t.opts.Escape(t.opts.WeeklyLabel),
		"{session_pct}", t.opts.Escape(five.percent()),
		"{session_reset}", t.opts.Escape(five.reset()),
		"{weekly_pct}", t.opts.Escape(seven.percent()),
		"{weekly_reset}", t.opts.Escape(seven.reset()),
		"{spend_pct}", t.opts.Escape(spend.percent()),
		"{spend_reset}", t.opts.Escape(spend.reset()),
		"{model}", t.opts.Escape(model),
		"{cost}", t.opts.Escape(cost),
		"{context_pct}", t.opts.Escape(context),
	).Replace(t.text)
}
