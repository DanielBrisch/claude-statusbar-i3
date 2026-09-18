package cli

import "fmt"

type VersionCommand struct {
	app *App
}

func NewVersionCommand(a *App) *VersionCommand {
	return &VersionCommand{app: a}
}

func (c *VersionCommand) Run([]string) (int, error) {
	fmt.Fprintln(c.app.Stdout, Version)
	return 0, nil
}
