package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/render"
	"github.com/DanielBrisch/claude-statusbar-i3/internal/setup"
	"github.com/DanielBrisch/claude-statusbar-i3/internal/state"
)

type DoctorCommand struct {
	app *App
}

func NewDoctorCommand(a *App) *DoctorCommand {
	return &DoctorCommand{app: a}
}

func (c *DoctorCommand) Run([]string) (int, error) {
	c.reportStatusLine()
	fmt.Fprintln(c.app.Stdout)
	return c.reportState()
}

func (c *DoctorCommand) reportStatusLine() {
	settings := c.app.paths().Settings()
	fmt.Fprintf(c.app.Stdout, "settings:   %s\n", settings)

	command, err := setup.NewInstaller(settings).Installed()
	switch {
	case err != nil:
		fmt.Fprintf(c.app.Stdout, "statusLine: unreadable (%v)\n", err)
	case command == "":
		fmt.Fprintln(c.app.Stdout, "statusLine: not configured")
		fmt.Fprintf(c.app.Stdout, "\nNothing feeds this tool until Claude Code runs it. Register it with:\n  %s init\n", BinaryName)
	case !strings.Contains(command, BinaryName):
		fmt.Fprintf(c.app.Stdout, "statusLine: runs something else\n            %s\n", command)
		fmt.Fprintf(c.app.Stdout, "\nThat command owns the slot, so nothing records your usage. Keep both with:\n  %s init --passthrough %q\n", BinaryName, command)
	default:
		fmt.Fprintln(c.app.Stdout, "statusLine: ok")
	}
}

func (c *DoctorCommand) reportState() (int, error) {
	store := c.app.store()
	path := store.Path()
	fmt.Fprintf(c.app.Stdout, "state file: %s\n", path)

	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(c.app.Stdout, "status:     missing")
		fmt.Fprintln(c.app.Stdout, "\nSend one message in Claude Code so its statusLine fires.")
		return 0, nil
	}

	if v, err := store.StoredVersion(); err == nil && v > state.Version {
		fmt.Fprintf(c.app.Stdout, "status:     written by a newer version (format %d, this binary reads %d)\n", v, state.Version)
		fmt.Fprintln(c.app.Stdout, "\nIt is being ignored rather than misread. Upgrade this binary, or delete the file.")
		return 0, nil
	}

	snapshot, err := store.Snapshot()
	if err != nil {
		fmt.Fprintln(c.app.Stdout, "status:     unreadable")
		return 0, err
	}

	fmt.Fprintf(c.app.Stdout, "written:    %s\n", info.ModTime().Format(time.RFC3339))
	if snapshot.FiveHour == nil && snapshot.SevenDay == nil {
		fmt.Fprintln(c.app.Stdout, "status:     no rate limits recorded")
		fmt.Fprintln(c.app.Stdout, "\nClaude Code omits rate_limits for API key, Bedrock and Vertex logins.")
		return 0, nil
	}

	fmt.Fprintln(c.app.Stdout, "status:     ok")
	fmt.Fprint(c.app.Stdout, "\n", render.NewRenderer(newRenderFlags().options(c.app.now())).Detail(snapshot).String())
	return 0, nil
}
