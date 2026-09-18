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

func New(out io.Writer) *Notifier {
	return &Notifier{
		Lookup: exec.LookPath,
		Run: func(name string, args ...string) error {
			return exec.Command(name, args...).Run()
		},
		Out: out,
	}
}

func (n *Notifier) Send(title, body string) error {
	for _, b := range n.backends(title, body) {
		if _, err := n.Lookup(b.name); err != nil {
			continue
		}
		if err := n.Run(b.name, b.args...); err == nil {
			return nil
		}
	}
	return n.printFallback(title, body)
}

func (n *Notifier) backends(title, body string) []backend {
	return []backend{
		newBackend("dunstify", "--appname", "claude-statusbar", "--replace", replaceID, title, body),
		newBackend("notify-send", "--app-name", "claude-statusbar", title, body),
	}
}

func (n *Notifier) printFallback(title, body string) error {
	out := n.Out
	if out == nil {
		out = os.Stdout
	}
	_, err := fmt.Fprintf(out, "%s\n%s", title, body)
	return err
}
