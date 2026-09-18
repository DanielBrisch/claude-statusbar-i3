package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/render"
)

type DoctorCommand struct {
	app *App
}

func NewDoctorCommand(a *App) *DoctorCommand {
	return &DoctorCommand{app: a}
}

func (c *DoctorCommand) Run([]string) (int, error) {
	path := c.app.paths().State()
	fmt.Fprintf(c.app.Stdout, "state file: %s\n", path)

	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(c.app.Stdout, "status:     missing")
		fmt.Fprintf(c.app.Stdout, "\nNothing has written it yet. Register the collector with:\n  %s init\n", BinaryName)
		fmt.Fprintln(c.app.Stdout, "then send one message in Claude Code so its statusLine fires.")
		return 0, nil
	}

	snapshot, err := c.app.store().Snapshot()
	if err != nil {
		fmt.Fprintln(c.app.Stdout, "status:     unreadable")
		return 0, err
	}

	fmt.Fprintf(c.app.Stdout, "written:    %s\n", info.ModTime().Format(time.RFC3339))
	if snapshot.FiveHour == nil && snapshot.SevenDay == nil {
		fmt.Fprintln(c.app.Stdout, "status:     no rate limits recorded")
		fmt.Fprintln(c.app.Stdout, "\nClaude Code omits rate_limits for API key, Bedrock and Vertex logins.")
		fmt.Fprintln(c.app.Stdout, "Check that the statusLine in your settings.json points at this binary.")
		return 0, nil
	}

	fmt.Fprintln(c.app.Stdout, "status:     ok")
	fmt.Fprint(c.app.Stdout, "\n", render.NewRenderer(newRenderFlags().options(c.app.now())).Detail(snapshot).String())
	return 0, nil
}
