package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/notify"
	"github.com/DanielBrisch/claude-usage-status-i3/internal/state"
)

const BinaryName = "claude-statusbar"

var Version = "dev"

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
	commands := newCommands(a)

	if len(a.Args) == 0 {
		commands.help().print(a.Stderr)
		return 2
	}

	name, rest := a.Args[0], a.Args[1:]
	command, ok := commands.lookup(name)
	if !ok {
		fmt.Fprintf(a.Stderr, "unknown subcommand %q\n\n", name)
		commands.help().print(a.Stderr)
		return 2
	}

	code, err := command.Run(rest)
	if err != nil {
		fmt.Fprintln(a.Stderr, err)
		return 1
	}
	return code
}

func (a *App) store() *state.Store {
	return state.New(a.paths().State())
}

func (a *App) paths() Paths {
	return NewPaths(a.StatePath, a.SettingsPath)
}

func (a *App) shell() *Shell {
	return NewShell(a.Shell)
}

func (a *App) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

func (a *App) env(key string) string {
	if a.Env != nil {
		return a.Env(key)
	}
	return os.Getenv(key)
}

func (a *App) notifier(out io.Writer) Notifier {
	if a.NewNotifier != nil {
		return a.NewNotifier(out)
	}
	return notify.New(out)
}

func (a *App) flagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	return fs
}
