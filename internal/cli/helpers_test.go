package cli

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var now = time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)

type fakeNotifier struct {
	title    string
	body     string
	sent     int
	out      io.Writer
	fallback bool
}

func (f *fakeNotifier) Send(title, body string) error {
	f.title, f.body = title, body
	f.sent++
	if f.fallback {
		fmt.Fprintf(f.out, "%s\n%s", title, body)
	}
	return nil
}

type harness struct {
	dir      string
	stdout   *bytes.Buffer
	stderr   *bytes.Buffer
	notifier *fakeNotifier
	env      map[string]string
	shell    []string
	shellErr error
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	return &harness{
		dir:      t.TempDir(),
		stdout:   &bytes.Buffer{},
		stderr:   &bytes.Buffer{},
		notifier: &fakeNotifier{},
		env:      map[string]string{},
	}
}

func (h *harness) run(stdin string, args ...string) int {
	h.stdout.Reset()
	h.stderr.Reset()
	app := &App{
		Args:         args,
		Stdin:        strings.NewReader(stdin),
		Stdout:       h.stdout,
		Stderr:       h.stderr,
		Env:          func(k string) string { return h.env[k] },
		Now:          func() time.Time { return now },
		StatePath:    filepath.Join(h.dir, "state.json"),
		SettingsPath: filepath.Join(h.dir, "settings.json"),
		NewNotifier: func(out io.Writer) Notifier {
			h.notifier.out = out
			return h.notifier
		},
		Shell: func(cmd string, stdin io.Reader, stdout, stderr io.Writer) error {
			h.shell = append(h.shell, cmd)
			if stdin != nil && stdout != nil {
				b, _ := io.ReadAll(stdin)
				fmt.Fprintf(stdout, "passthrough saw %d bytes", len(b))
			}
			return h.shellErr
		},
	}
	return app.Run()
}

func payloadJSON(fiveHourPct float64) string {
	return fmt.Sprintf(`{
	  "session_id":"s1",
	  "model":{"display_name":"Opus"},
	  "cost":{"total_cost_usd":2.41,"total_duration_ms":2700000},
	  "context_window":{"used_percentage":18,"context_window_size":200000},
	  "rate_limits":{
	    "five_hour":{"used_percentage":%v,"resets_at":%d},
	    "seven_day":{"used_percentage":21,"resets_at":%d}
	  }
	}`, fiveHourPct, now.Add(102*time.Minute).Unix(), now.Add(100*time.Hour).Unix())
}
