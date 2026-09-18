package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/notify"
	"github.com/DanielBrisch/claude-usage-status-i3/internal/render"
	"github.com/DanielBrisch/claude-usage-status-i3/internal/state"
	"github.com/DanielBrisch/claude-usage-status-i3/internal/statusline"
	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

var Version = "dev"

const (
	notificationTitle = "Claude usage"
	i3blocksUrgent    = 33
)

type Notifier interface {
	Send(title, body string) error
}

type App struct {
	Args         []string
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
	Env          func(string) string
	Now          func() time.Time
	StatePath    string
	SettingsPath string
	NewNotifier  func(out io.Writer) Notifier
	Shell        func(cmd string, stdin io.Reader, stdout, stderr io.Writer) error
}

func (a *App) Run() int {
	if len(a.Args) == 0 {
		a.usage()
		return 2
	}
	sub, rest := a.Args[0], a.Args[1:]

	var err error
	code := 0
	switch sub {
	case "collect":
		err = a.collect(rest)
	case "render":
		code, err = a.render(rest)
	case "detail":
		err = a.detail(rest)
	case "init":
		err = a.init(rest)
	case "doctor":
		err = a.doctor()
	case "version", "--version", "-v":
		fmt.Fprintln(a.Stdout, Version)
	case "help", "--help", "-h":
		a.usage()
	default:
		fmt.Fprintf(a.Stderr, "unknown subcommand %q\n\n", sub)
		a.usage()
		return 2
	}

	if err != nil {
		fmt.Fprintln(a.Stderr, err)
		return 1
	}
	return code
}

func (a *App) collect(args []string) error {
	fs := a.flags("collect")
	print := fs.String("print", "compact", "what to echo back to Claude Code: compact|none")
	passthrough := fs.String("passthrough", "", "shell command to receive the same payload and own the status line")
	refresh := fs.String("refresh", "", "shell command to run after recording, to poke your bar (e.g. 'pkill -SIGRTMIN+12 i3blocks')")
	o := a.renderFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	raw, err := io.ReadAll(a.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	payload, err := statusline.Parse(strings.NewReader(string(raw)))
	if err != nil {
		return err
	}

	store := state.New(a.statePath())
	if err := store.Merge(payload, a.now()); err != nil {
		return err
	}

	if *refresh != "" {
		if err := a.shell(*refresh, nil, io.Discard, io.Discard); err != nil {
			fmt.Fprintf(a.Stderr, "refresh command failed: %v\n", err)
		}
	}

	if *passthrough != "" {
		return a.shell(*passthrough, strings.NewReader(string(raw)), a.Stdout, a.Stderr)
	}
	if *print == "none" {
		return nil
	}

	snap, err := store.Snapshot()
	if err != nil {
		return err
	}
	if text := render.Compact(snap, a.options(o)).FullText; text != "" {
		fmt.Fprintln(a.Stdout, text)
	}
	return nil
}

func (a *App) render(args []string) (int, error) {
	fs := a.flags("render")
	format := fs.String("format", "i3blocks", "i3blocks|waybar|polybar|plain|json")
	o := a.renderFlags(fs)
	if err := fs.Parse(args); err != nil {
		return 0, err
	}

	snap, err := state.New(a.statePath()).Snapshot()
	if err != nil {
		return 0, err
	}
	opts := a.options(o)

	var out string
	switch *format {
	case "i3blocks":
		out = render.I3blocks(snap, opts)
	case "waybar":
		out = render.Waybar(snap, opts)
	case "polybar":
		out = render.Polybar(snap, opts)
	case "plain":
		out = render.Plain(snap, opts)
	case "json":
		out = render.JSON(snap, opts)
	default:
		return 0, fmt.Errorf("unknown format %q (want i3blocks, waybar, polybar, plain or json)", *format)
	}

	if out != "" {
		fmt.Fprint(a.Stdout, out)
		if !strings.HasSuffix(out, "\n") {
			fmt.Fprintln(a.Stdout)
		}
	}
	if *format == "i3blocks" && render.Compact(snap, opts).Level == usage.LevelUrgent {
		return i3blocksUrgent, nil
	}
	return 0, nil
}

func (a *App) detail(args []string) error {
	fs := a.flags("detail")
	send := fs.Bool("notify", false, "send the detail as a desktop notification instead of printing it")
	o := a.renderFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	snap, err := state.New(a.statePath()).Snapshot()
	if err != nil {
		return err
	}
	body := render.Detail(snap, a.options(o))
	if *send {
		return a.notifier(a.Stdout).Send(notificationTitle, body)
	}
	fmt.Fprint(a.Stdout, body)
	return nil
}

func (a *App) init(args []string) error {
	fs := a.flags("init")
	bar := fs.String("bar", "i3blocks", "which bar snippet to print: i3blocks|waybar|polybar")
	force := fs.Bool("force", false, "replace an existing statusLine")
	passthrough := fs.String("passthrough", "", "keep an existing status line by chaining it")
	if err := fs.Parse(args); err != nil {
		return err
	}

	command := binaryName() + " collect"
	if *passthrough != "" {
		command = fmt.Sprintf("%s collect --passthrough %q", binaryName(), *passthrough)
	}
	if err := installStatusLine(a.settingsPath(), command, *force); err != nil {
		return err
	}

	fmt.Fprintf(a.Stdout, "statusLine written to %s\n\n", a.settingsPath())
	fmt.Fprint(a.Stdout, snippet(*bar))
	return nil
}

func (a *App) doctor() error {
	path := a.statePath()
	fmt.Fprintf(a.Stdout, "state file: %s\n", path)

	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(a.Stdout, "status:     missing")
		fmt.Fprintf(a.Stdout, "\nNothing has written it yet. Register the collector with:\n  %s init\n", binaryName())
		fmt.Fprintln(a.Stdout, "then send one message in Claude Code so its statusLine fires.")
		return nil
	}

	snap, err := state.New(path).Snapshot()
	if err != nil {
		fmt.Fprintln(a.Stdout, "status:     unreadable")
		return err
	}

	fmt.Fprintf(a.Stdout, "written:    %s\n", info.ModTime().Format(time.RFC3339))
	if snap.FiveHour == nil && snap.SevenDay == nil {
		fmt.Fprintln(a.Stdout, "status:     no rate limits recorded")
		fmt.Fprintln(a.Stdout, "\nClaude Code omits rate_limits for API key, Bedrock and Vertex logins.")
		fmt.Fprintln(a.Stdout, "Check that the statusLine in your settings.json points at this binary.")
		return nil
	}

	fmt.Fprintln(a.Stdout, "status:     ok")
	fmt.Fprint(a.Stdout, "\n", render.Detail(snap, a.options(defaultRenderOptions())))
	return nil
}

type renderOptions struct {
	label       *string
	weeklyLabel *string
	template    *string
	markup      *string
	iconSize    *string
	warn        *float64
	crit        *float64
	urgent      *float64
	colorWarn   *string
	colorCrit   *string
}

func (a *App) renderFlags(fs *flag.FlagSet) *renderOptions {
	d := usage.DefaultThresholds()
	c := render.DefaultColors()
	return &renderOptions{
		label:       fs.String("label", render.DefaultLabel, "prefix for the session segment"),
		weeklyLabel: fs.String("weekly-label", render.DefaultWeeklyLabel, "prefix for the weekly segment"),
		template:    fs.String("template", "", "custom layout, e.g. '{label} {session_pct} {session_reset}'"),
		markup:      fs.String("markup", string(render.MarkupNone), "none|pango; pango lets the icon be enlarged, and your bar must be told to parse it"),
		iconSize:    fs.String("icon-size", render.DefaultIconSize, "pango size for the icon when --markup pango: small, medium, large, x-large, xx-large"),
		warn:        fs.Float64("warn", d.Warn, "percentage that turns the block yellow"),
		crit:        fs.Float64("crit", d.Crit, "percentage that turns the block red"),
		urgent:      fs.Float64("urgent", d.Urgent, "percentage that marks the block urgent"),
		colorWarn:   fs.String("color-warn", c.Warn, "hex colour for the warn level"),
		colorCrit:   fs.String("color-crit", c.Crit, "hex colour for the crit level"),
	}
}

func defaultRenderOptions() *renderOptions {
	d := usage.DefaultThresholds()
	c := render.DefaultColors()
	label, weekly, empty := render.DefaultLabel, render.DefaultWeeklyLabel, ""
	markup, iconSize := string(render.MarkupNone), render.DefaultIconSize
	return &renderOptions{
		label: &label, weeklyLabel: &weekly, template: &empty,
		markup: &markup, iconSize: &iconSize,
		warn: &d.Warn, crit: &d.Crit, urgent: &d.Urgent,
		colorWarn: &c.Warn, colorCrit: &c.Crit,
	}
}

func (a *App) options(o *renderOptions) render.Options {
	return render.Options{
		Now:         a.now(),
		Label:       *o.label,
		WeeklyLabel: *o.weeklyLabel,
		Template:    *o.template,
		Markup:      render.Markup(*o.markup),
		IconSize:    *o.iconSize,
		Thresholds:  usage.Thresholds{Warn: *o.warn, Crit: *o.crit, Urgent: *o.urgent},
		Colors:      render.Colors{Warn: *o.colorWarn, Crit: *o.colorCrit},
	}
}

func (a *App) flags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	return fs
}

func (a *App) usage() {
	fmt.Fprintf(a.Stderr, `%s %s — Claude Code usage on your status bar

Usage:
  %[1]s collect [--print compact|none] [--passthrough CMD] [--refresh CMD]
        Reads the Claude Code statusLine payload on stdin and records it.
  %[1]s render [--format i3blocks|waybar|polybar|plain|json]
        Renders the recorded usage for your bar.
  %[1]s detail [--notify]
        Prints or notifies the full breakdown.
  %[1]s init [--bar i3blocks|waybar|polybar] [--force] [--passthrough CMD]
        Registers the collector in Claude Code settings and prints the bar snippet.
  %[1]s doctor
        Explains why the bar might be empty.

Run "%[1]s <subcommand> --help" for the flags of each one.
`, binaryName(), Version)
}

func (a *App) shell(cmd string, stdin io.Reader, stdout, stderr io.Writer) error {
	if a.Shell != nil {
		return a.Shell(cmd, stdin, stdout, stderr)
	}
	c := exec.Command("sh", "-c", cmd)
	c.Stdin = stdin
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}

func (a *App) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

func (a *App) env(k string) string {
	if a.Env != nil {
		return a.Env(k)
	}
	return os.Getenv(k)
}

func (a *App) notifier(out io.Writer) Notifier {
	if a.NewNotifier != nil {
		return a.NewNotifier(out)
	}
	return notify.New(out)
}

func (a *App) statePath() string {
	if a.StatePath != "" {
		return a.StatePath
	}
	return DefaultStatePath()
}

func (a *App) settingsPath() string {
	if a.SettingsPath != "" {
		return a.SettingsPath
	}
	return DefaultSettingsPath()
}

func DefaultStatePath() string {
	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return filepath.Join(dir, "claude-statusbar", "state.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "claude-statusbar", "state.json")
	}
	return filepath.Join(home, ".local", "state", "claude-statusbar", "state.json")
}

func DefaultSettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".claude/settings.json"
	}
	return filepath.Join(home, ".claude", "settings.json")
}

func binaryName() string { return "claude-statusbar" }
