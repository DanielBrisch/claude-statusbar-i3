package cli

import (
	"flag"
	"time"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/render"
	"github.com/DanielBrisch/claude-statusbar-i3/internal/usage"
)

type renderFlags struct {
	label       string
	weeklyLabel string
	template    string
	markup      string
	iconSize    string
	warn        float64
	crit        float64
	urgent      float64
	colorWarn   string
	colorCrit   string
	urgentExit  bool
}

func newRenderFlags() *renderFlags {
	d := usage.DefaultThresholds()
	c := render.DefaultColors()
	return &renderFlags{
		label:       render.DefaultLabel,
		weeklyLabel: render.DefaultWeeklyLabel,
		markup:      string(render.MarkupNone),
		iconSize:    render.DefaultIconSize,
		warn:        d.Warn,
		crit:        d.Crit,
		urgent:      d.Urgent,
		colorWarn:   c.Warn,
		colorCrit:   c.Crit,
	}
}

func (f *renderFlags) bind(fs *flag.FlagSet) {
	fs.StringVar(&f.label, "label", f.label, "prefix for the session segment")
	fs.StringVar(&f.weeklyLabel, "weekly-label", f.weeklyLabel, "prefix for the weekly segment")
	fs.StringVar(&f.template, "template", f.template, "custom layout, e.g. '{label} {session_pct} {session_reset}'")
	fs.StringVar(&f.markup, "markup", f.markup, "none|pango; pango lets the icon be enlarged, and your bar must be told to parse it")
	fs.StringVar(&f.iconSize, "icon-size", f.iconSize, "pango size for the icon when --markup pango: small, medium, large, x-large, xx-large")
	fs.Float64Var(&f.warn, "warn", f.warn, "percentage that reaches the warn level")
	fs.Float64Var(&f.crit, "crit", f.crit, "percentage that reaches the crit level")
	fs.Float64Var(&f.urgent, "urgent", f.urgent, "percentage that reaches the urgent level")
	fs.StringVar(&f.colorWarn, "color-warn", f.colorWarn, "hex colour once --warn is crossed; empty keeps your bar's own text colour")
	fs.StringVar(&f.colorCrit, "color-crit", f.colorCrit, "hex colour once --crit is crossed; empty keeps your bar's own text colour")
	fs.BoolVar(&f.urgentExit, "urgent-exit", f.urgentExit, "exit 33 past --urgent so i3blocks marks the block urgent, which recolours it")
}

func (f *renderFlags) options(now time.Time) render.Options {
	o := render.NewOptions(now)
	o.Label = f.label
	o.WeeklyLabel = f.weeklyLabel
	o.Template = f.template
	o.Markup = render.Markup(f.markup)
	o.IconSize = f.iconSize
	o.UrgentExit = f.urgentExit
	o.Thresholds = usage.NewThresholds(f.warn, f.crit, f.urgent)
	o.Colors = render.NewColors(f.colorWarn, f.colorCrit)
	return o
}
