package cli

import (
	"fmt"
	"strings"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/render"
)

type RenderCommand struct {
	app *App
}

func NewRenderCommand(a *App) *RenderCommand {
	return &RenderCommand{app: a}
}

func (c *RenderCommand) Run(args []string) (int, error) {
	fs := c.app.flagSet("render")
	format := fs.String("format", "i3blocks", "i3blocks|waybar|polybar|plain|json")
	flags := newRenderFlags()
	flags.bind(fs)
	if err := fs.Parse(args); err != nil {
		return 0, err
	}

	snapshot, err := c.app.store().Snapshot()
	if err != nil {
		return 0, err
	}

	formatter, err := render.NewRegistry(flags.options(c.app.now())).Lookup(*format)
	if err != nil {
		return 0, err
	}

	out := formatter.Render(snapshot)
	if !out.Empty() {
		fmt.Fprint(c.app.Stdout, out.Text)
		if !strings.HasSuffix(out.Text, "\n") {
			fmt.Fprintln(c.app.Stdout)
		}
	}
	return out.ExitCode, nil
}
