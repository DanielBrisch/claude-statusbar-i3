package render

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

const (
	Placeholder        = "—"
	DefaultLabel       = "✳"
	DefaultWeeklyLabel = "week"
)

const DetailCommand = "claude-statusbar detail --notify"

type Markup string

const (
	MarkupNone      Markup = "none"
	MarkupPango     Markup = "pango"
	DefaultIconSize        = "large"
)

type Colors struct {
	OK   string
	Warn string
	Crit string
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

type Options struct {
	Now         time.Time
	Label       string
	WeeklyLabel string
	Template    string
	Markup      Markup
	IconSize    string
	Thresholds  usage.Thresholds
	Colors      Colors
}

func (o Options) icon(label string) string {
	if o.Markup != MarkupPango || label == "" {
		return label
	}
	size := o.IconSize
	if size == "" {
		size = DefaultIconSize
	}
	return fmt.Sprintf("<span size=%q>%s</span>", size, html.EscapeString(label))
}

func (o Options) escape(s string) string {
	if o.Markup != MarkupPango {
		return s
	}
	return html.EscapeString(s)
}

func (o Options) withoutMarkup() Options {
	o.Markup = MarkupNone
	return o
}

type Result struct {
	FullText   string
	ShortText  string
	Color      string
	Level      usage.Level
	Percentage float64
}

func Compact(s usage.Snapshot, o Options) Result {
	if !s.HasVisibleWindow(o.Now) {
		return Result{}
	}

	level := s.Level(o.Now, o.Thresholds)
	r := Result{
		Level:      level,
		Color:      o.Colors.For(level),
		Percentage: primaryPercentage(s, o.Now),
		ShortText:  pct(s.FiveHour, o.Now) + "│" + pct(s.SevenDay, o.Now),
	}

	if o.Template != "" {
		r.FullText = expand(o.Template, s, o)
		return r
	}

	r.FullText = strings.TrimSpace(join(
		segment(o.icon(o.Label), s.FiveHour, o),
		o.escape("│"),
		segment(o.escape(o.WeeklyLabel), s.SevenDay, o),
	))
	return r
}

func I3blocks(s usage.Snapshot, o Options) string {
	r := Compact(s, o)
	if r.FullText == "" {
		return ""
	}
	out := r.FullText + "\n" + r.ShortText + "\n"
	if r.Color != "" {
		out += r.Color + "\n"
	}
	return out
}

func Waybar(s usage.Snapshot, o Options) string {
	r := Compact(s, o)
	if r.FullText == "" {
		return "{}"
	}
	b, err := json.Marshal(struct {
		Text       string  `json:"text"`
		Tooltip    string  `json:"tooltip"`
		Class      string  `json:"class"`
		Percentage float64 `json:"percentage"`
	}{
		Text:       r.FullText,
		Tooltip:    Detail(s, o),
		Class:      r.Level.String(),
		Percentage: r.Percentage,
	})
	if err != nil {
		return "{}"
	}
	return string(b)
}

func Polybar(s usage.Snapshot, o Options) string {
	r := Compact(s, o.withoutMarkup())
	if r.FullText == "" {
		return ""
	}
	text := r.FullText
	if r.Color != "" {
		text = "%{F" + r.Color + "}" + text + "%{F-}"
	}
	return "%{A1:" + DetailCommand + ":}" + text + "%{A}"
}

func Plain(s usage.Snapshot, o Options) string {
	return Compact(s, o.withoutMarkup()).FullText
}

func JSON(s usage.Snapshot, o Options) string {
	o = o.withoutMarkup()
	r := Compact(s, o)
	b, err := json.MarshalIndent(struct {
		Text       string   `json:"text"`
		Level      string   `json:"level"`
		Percentage float64  `json:"percentage"`
		FiveHour   *window  `json:"five_hour"`
		SevenDay   *window  `json:"seven_day"`
		SpendLimit *window  `json:"spend_limit"`
		Session    *session `json:"session"`
		UpdatedAt  string   `json:"updated_at"`
	}{
		Text:       r.FullText,
		Level:      r.Level.String(),
		Percentage: r.Percentage,
		FiveHour:   toWindow(s.FiveHour, o.Now),
		SevenDay:   toWindow(s.SevenDay, o.Now),
		SpendLimit: toWindow(s.SpendLimit, o.Now),
		Session:    toSession(s.Session),
		UpdatedAt:  stamp(s.UpdatedAt),
	}, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

type window struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       string  `json:"resets_at"`
	TimeLeft       string  `json:"time_left"`
	Expired        bool    `json:"expired"`
}

type session struct {
	Model             string  `json:"model"`
	CostUSD           float64 `json:"cost_usd"`
	ContextUsedPct    float64 `json:"context_used_pct"`
	ContextWindowSize int64   `json:"context_window_size"`
	DurationMinutes   int     `json:"duration_minutes"`
}

func Detail(s usage.Snapshot, o Options) string {
	var b strings.Builder
	b.WriteString(detailLine("5h window", s.FiveHour, o.Now))
	b.WriteString(detailLine("Weekly", s.SevenDay, o.Now))
	if s.SpendLimit != nil {
		b.WriteString(detailLine("Spend", s.SpendLimit, o.Now))
	}
	fmt.Fprintf(&b, "as of %s\n", stampTime(s.UpdatedAt))
	return b.String()
}

func detailLine(label string, l *usage.Limit, now time.Time) string {
	if l == nil || l.Expired(now) {
		return fmt.Sprintf("%-11s %5s\n", label, Placeholder)
	}
	return fmt.Sprintf("%-11s %5s  resets %s (%s)\n", label,
		fmt.Sprintf("%.0f%%", l.UsedPercentage),
		l.ResetsAt.Format("Jan 02 15:04"),
		usage.FormatTimeLeft(l.TimeLeft(now)))
}

func segment(label string, l *usage.Limit, o Options) string {
	body := Placeholder
	if l != nil && !l.Expired(o.Now) {
		body = fmt.Sprintf("%.0f%% %s", l.UsedPercentage, usage.FormatTimeLeft(l.TimeLeft(o.Now)))
	}
	body = o.escape(body)
	if label == "" {
		return body
	}
	return label + " " + body
}

func join(parts ...string) string {
	return strings.Join(parts, " ")
}

func pct(l *usage.Limit, now time.Time) string {
	if l == nil || l.Expired(now) {
		return Placeholder
	}
	return fmt.Sprintf("%.0f%%", l.UsedPercentage)
}

func reset(l *usage.Limit, now time.Time) string {
	if l == nil || l.Expired(now) {
		return Placeholder
	}
	return usage.FormatTimeLeft(l.TimeLeft(now))
}

func primaryPercentage(s usage.Snapshot, now time.Time) float64 {
	if s.FiveHour != nil && !s.FiveHour.Expired(now) {
		return s.FiveHour.UsedPercentage
	}
	if s.SevenDay != nil && !s.SevenDay.Expired(now) {
		return s.SevenDay.UsedPercentage
	}
	return 0
}

func expand(tmpl string, s usage.Snapshot, o Options) string {
	model, cost, context := "", "", ""
	if s.Session != nil {
		model = s.Session.Model
		cost = fmt.Sprintf("$%.2f", s.Session.CostUSD)
		context = fmt.Sprintf("%.0f%%", s.Session.ContextUsedPct)
	}
	return strings.NewReplacer(
		"{label}", o.icon(o.Label),
		"{weekly_label}", o.escape(o.WeeklyLabel),
		"{session_pct}", o.escape(pct(s.FiveHour, o.Now)),
		"{session_reset}", o.escape(reset(s.FiveHour, o.Now)),
		"{weekly_pct}", o.escape(pct(s.SevenDay, o.Now)),
		"{weekly_reset}", o.escape(reset(s.SevenDay, o.Now)),
		"{spend_pct}", o.escape(pct(s.SpendLimit, o.Now)),
		"{spend_reset}", o.escape(reset(s.SpendLimit, o.Now)),
		"{model}", o.escape(model),
		"{cost}", o.escape(cost),
		"{context_pct}", o.escape(context),
	).Replace(tmpl)
}

func toWindow(l *usage.Limit, now time.Time) *window {
	if l == nil {
		return nil
	}
	return &window{
		UsedPercentage: l.UsedPercentage,
		ResetsAt:       stamp(l.ResetsAt),
		TimeLeft:       usage.FormatTimeLeft(l.TimeLeft(now)),
		Expired:        l.Expired(now),
	}
}

func toSession(s *usage.Session) *session {
	if s == nil {
		return nil
	}
	return &session{
		Model:             s.Model,
		CostUSD:           s.CostUSD,
		ContextUsedPct:    s.ContextUsedPct,
		ContextWindowSize: s.ContextWindowSize,
		DurationMinutes:   int(s.Duration / time.Minute),
	}
}

func stamp(t time.Time) string {
	if t.IsZero() || t.Unix() <= 0 {
		return ""
	}
	return t.Format(time.RFC3339)
}

func stampTime(t time.Time) string {
	if t.IsZero() || t.Unix() <= 0 {
		return "?"
	}
	return t.Format("15:04")
}
