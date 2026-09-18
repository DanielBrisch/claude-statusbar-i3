package cli

import (
	"fmt"
	"io"
)

type HelpCommand struct {
	app *App
}

func NewHelpCommand(a *App) *HelpCommand {
	return &HelpCommand{app: a}
}

func (c *HelpCommand) Run([]string) (int, error) {
	c.print(c.app.Stdout)
	return 0, nil
}

func (c *HelpCommand) print(w io.Writer) {
	fmt.Fprintf(w, `%s %s — Claude Code usage on your status bar

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
`, BinaryName, Version)
}
