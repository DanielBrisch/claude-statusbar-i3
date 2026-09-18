package cli

import (
	"io"
	"os/exec"
)

type Shell struct {
	run func(cmd string, stdin io.Reader, stdout, stderr io.Writer) error
}

func NewShell(run func(cmd string, stdin io.Reader, stdout, stderr io.Writer) error) *Shell {
	return &Shell{run: run}
}

func (s *Shell) Run(cmd string, stdin io.Reader, stdout, stderr io.Writer) error {
	if s != nil && s.run != nil {
		return s.run(cmd, stdin, stdout, stderr)
	}
	c := exec.Command("sh", "-c", cmd)
	c.Stdin = stdin
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}
