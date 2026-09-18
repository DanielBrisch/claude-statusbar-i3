package main

import (
	"os"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/cli"
)

func main() {
	app := &cli.App{
		Args:   os.Args[1:],
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	os.Exit(app.Run())
}
