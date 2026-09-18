package cli

import (
	"fmt"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/setup"
)

type InitCommand struct {
	app *App
}

func NewInitCommand(a *App) *InitCommand {
	return &InitCommand{app: a}
}

func (c *InitCommand) Run(args []string) (int, error) {
	fs := c.app.flagSet("init")
	bar := fs.String("bar", "i3blocks", "which bar snippet to print: i3blocks|waybar|polybar")
	force := fs.Bool("force", false, "replace an existing statusLine")
	passthrough := fs.String("passthrough", "", "keep an existing status line by chaining it")
	if err := fs.Parse(args); err != nil {
		return 0, err
	}

	command := BinaryName + " collect"
	if *passthrough != "" {
		command = fmt.Sprintf("%s collect --passthrough %q", BinaryName, *passthrough)
	}

	installer := setup.NewInstaller(c.app.paths().Settings())
	if err := installer.Install(command, *force); err != nil {
		return 0, err
	}

	fmt.Fprintf(c.app.Stdout, "statusLine written to %s\n\n", installer.Path())
	fmt.Fprint(c.app.Stdout, setup.NewSnippet(*bar, BinaryName).String())
	return 0, nil
}
