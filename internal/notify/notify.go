package notify

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

const replaceID = "9471"

type Notifier struct {
	Lookup func(string) (string, error)
	Run    func(name string, args ...string) error
	Out    io.Writer
}

func New() *Notifier {
	return &Notifier{
		Lookup: exec.LookPath,
		Run: func(name string, args ...string) error {
			return exec.Command(name, args...).Run()
		},
		Out: os.Stdout,
	}
}

func (n *Notifier) Send(title, body string) error {
	candidates := []struct {
		name string
		args []string
	}{
		{"dunstify", []string{"--appname", "claude-statusbar", "--replace", replaceID, title, body}},
		{"notify-send", []string{"--app-name", "claude-statusbar", title, body}},
	}

	for _, c := range candidates {
		if _, err := n.Lookup(c.name); err != nil {
			continue
		}
		if err := n.Run(c.name, c.args...); err == nil {
			return nil
		}
	}

	out := n.Out
	if out == nil {
		out = os.Stdout
	}
	_, err := fmt.Fprintf(out, "%s\n%s", title, body)
	return err
}
