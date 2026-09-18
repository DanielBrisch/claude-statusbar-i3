package cli

import (
	"fmt"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/render"
)

const notificationTitle = "Claude usage"

type DetailCommand struct {
	app *App
}

func NewDetailCommand(a *App) *DetailCommand {
	return &DetailCommand{app: a}
}

func (c *DetailCommand) Run(args []string) (int, error) {
	fs := c.app.flagSet("detail")
	send := fs.Bool("notify", false, "send the detail as a desktop notification instead of printing it")
	flags := newRenderFlags()
	flags.bind(fs)
	if err := fs.Parse(args); err != nil {
		return 0, err
	}

	snapshot, err := c.app.store().Snapshot()
	if err != nil {
		return 0, err
	}
	body := render.NewRenderer(flags.options(c.app.now())).Detail(snapshot).String()

	if *send {
		return 0, c.app.notifier(c.app.Stdout).Send(notificationTitle, body)
	}
	fmt.Fprint(c.app.Stdout, body)
	return 0, nil
}
