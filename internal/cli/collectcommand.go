package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/render"
	"github.com/DanielBrisch/claude-statusbar-i3/internal/statusline"
)

type CollectCommand struct {
	app *App
}

func NewCollectCommand(a *App) *CollectCommand {
	return &CollectCommand{app: a}
}

func (c *CollectCommand) Run(args []string) (int, error) {
	fs := c.app.flagSet("collect")
	print := fs.String("print", "compact", "what to echo back to Claude Code: compact|none")
	passthrough := fs.String("passthrough", "", "shell command to receive the same payload and own the status line")
	refresh := fs.String("refresh", "", "shell command to run after recording, to poke your bar (e.g. 'pkill -SIGRTMIN+12 i3blocks')")
	flags := newRenderFlags()
	flags.bind(fs)
	if err := fs.Parse(args); err != nil {
		return 0, err
	}

	raw, err := io.ReadAll(c.app.Stdin)
	if err != nil {
		return 0, fmt.Errorf("read stdin: %w", err)
	}
	payload, err := statusline.NewParser().Parse(strings.NewReader(string(raw)))
	if err != nil {
		return 0, err
	}

	store := c.app.store()
	if err := store.Merge(payload, c.app.now()); err != nil {
		return 0, err
	}

	if *refresh != "" {
		if err := c.app.shell().Run(*refresh, nil, io.Discard, io.Discard); err != nil {
			fmt.Fprintf(c.app.Stderr, "refresh command failed: %v\n", err)
		}
	}
	if *passthrough != "" {
		return 0, c.app.shell().Run(*passthrough, strings.NewReader(string(raw)), c.app.Stdout, c.app.Stderr)
	}
	if *print == "none" {
		return 0, nil
	}

	snapshot, err := store.Snapshot()
	if err != nil {
		return 0, err
	}
	block := render.NewRenderer(flags.options(c.app.now())).Block(snapshot)
	if !block.Empty() {
		fmt.Fprintln(c.app.Stdout, block.FullText)
	}
	return 0, nil
}
