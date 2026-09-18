package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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

func TestCollectStoresStateAndEchoesTheCompactLine(t *testing.T) {
	h := newHarness(t)

	if code := h.run(payloadJSON(63), "collect"); code != 0 {
		t.Fatalf("collect exit = %d, stderr = %s", code, h.stderr)
	}
	if got, want := strings.TrimSpace(h.stdout.String()), "✳ 63% 1h42 │ week 21% 4d"; got != want {
		t.Errorf("collect stdout = %q, want %q", got, want)
	}

	if code := h.run("", "render", "--format", "plain"); code != 0 {
		t.Fatalf("render exit = %d, stderr = %s", code, h.stderr)
	}
	if got, want := strings.TrimSpace(h.stdout.String()), "✳ 63% 1h42 │ week 21% 4d"; got != want {
		t.Errorf("render stdout = %q, want %q", got, want)
	}
}

func TestCollectPrintNoneStillWritesState(t *testing.T) {
	h := newHarness(t)

	if code := h.run(payloadJSON(63), "collect", "--print", "none"); code != 0 {
		t.Fatalf("collect exit = %d, stderr = %s", code, h.stderr)
	}
	if h.stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", h.stdout)
	}
	if _, err := os.Stat(filepath.Join(h.dir, "state.json")); err != nil {
		t.Errorf("state was not written: %v", err)
	}
}

func TestCollectSurvivesAPayloadWithoutRateLimits(t *testing.T) {
	h := newHarness(t)

	code := h.run(`{"session_id":"s1","model":{"display_name":"Opus"}}`, "collect")
	if code != 0 {
		t.Fatalf("collect exit = %d, stderr = %s", code, h.stderr)
	}
	if h.stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty when there are no rate limits", h.stdout)
	}
}

func TestCollectReportsMalformedInput(t *testing.T) {
	h := newHarness(t)

	if code := h.run("{not json", "collect"); code == 0 {
		t.Fatal("collect of malformed JSON exited 0, want non-zero")
	}
	if !strings.Contains(h.stderr.String(), "decode") {
		t.Errorf("stderr = %q, want it to explain the decode failure", h.stderr)
	}
}

func TestRenderOnAnEmptyStateIsSilentAndSuccessful(t *testing.T) {
	h := newHarness(t)

	if code := h.run("", "render", "--format", "i3blocks"); code != 0 {
		t.Errorf("render exit = %d, want 0 — a missing state must never break the bar", code)
	}
	if h.stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", h.stdout)
	}
}

func TestRenderExitsThirtyThreeWhenUrgent(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(97), "collect", "--print", "none")

	if code := h.run("", "render", "--format", "i3blocks"); code != 33 {
		t.Errorf("render exit = %d, want 33 (i3blocks urgent)", code)
	}
}

func TestRenderDoesNotExitThirtyThreeForOtherFormats(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(97), "collect", "--print", "none")

	if code := h.run("", "render", "--format", "waybar"); code != 0 {
		t.Errorf("render exit = %d, want 0 — exit 33 is an i3blocks convention", code)
	}
}

func TestRenderWaybarIsValidJSON(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(63), "collect", "--print", "none")
	h.run("", "render", "--format", "waybar")

	var out map[string]any
	if err := json.Unmarshal(h.stdout.Bytes(), &out); err != nil {
		t.Fatalf("waybar output is not JSON: %v\n%s", err, h.stdout)
	}
	if out["text"] != "✳ 63% 1h42 │ week 21% 4d" {
		t.Errorf("text = %v", out["text"])
	}
}

func TestClickingTheI3blocksBlockDoesNothing(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(63), "collect", "--print", "none")

	for _, button := range []string{"1", "2", "3", "4", "5"} {
		h.env["BLOCK_BUTTON"] = button
		if code := h.run("", "render", "--format", "i3blocks"); code != 0 {
			t.Errorf("button %s: exit = %d", button, code)
		}
	}
	if h.notifier.sent != 0 {
		t.Errorf("notifier called %d times, want 0 — the icon makes the block self-explanatory, so the click no longer pops anything", h.notifier.sent)
	}
	if !strings.HasPrefix(h.stdout.String(), "\u2733 63%") {
		t.Errorf("the block must still render normally, got %q", h.stdout)
	}
}

func TestDetailNotifyUsesTheNotifier(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(63), "collect", "--print", "none")

	if code := h.run("", "detail", "--notify"); code != 0 {
		t.Fatalf("detail exit = %d, stderr = %s", code, h.stderr)
	}
	if h.notifier.sent != 1 {
		t.Errorf("notifier called %d times, want 1", h.notifier.sent)
	}
}

func TestDetailWithoutNotifyPrintsToStdout(t *testing.T) {
	h := newHarness(t)
	h.run(payloadJSON(63), "collect", "--print", "none")

	h.run("", "detail")
	if !strings.Contains(h.stdout.String(), "5h window") {
		t.Errorf("stdout = %q, want the detail", h.stdout)
	}
	if h.notifier.sent != 0 {
		t.Errorf("notifier called %d times without --notify, want 0", h.notifier.sent)
	}
}

func TestInitWritesTheStatusLineIntoSettings(t *testing.T) {
	h := newHarness(t)

	if code := h.run("", "init"); code != 0 {
		t.Fatalf("init exit = %d, stderr = %s", code, h.stderr)
	}
	b, err := os.ReadFile(filepath.Join(h.dir, "settings.json"))
	if err != nil {
		t.Fatalf("settings not written: %v", err)
	}
	if !strings.Contains(string(b), "collect") {
		t.Errorf("settings = %s, want a statusLine running collect", b)
	}
	if !strings.Contains(h.stdout.String(), "i3blocks") {
		t.Errorf("init should print the bar snippet, got: %s", h.stdout)
	}
}

func TestDoctorExplainsAnEmptyState(t *testing.T) {
	h := newHarness(t)

	h.run("", "doctor")
	out := h.stdout.String()
	if !strings.Contains(out, "state.json") {
		t.Errorf("doctor output = %q, want the state path", out)
	}
	if !strings.Contains(strings.ToLower(out), "statusline") {
		t.Errorf("doctor output = %q, want it to point at the statusLine setup", out)
	}
}

func TestUnknownSubcommandFails(t *testing.T) {
	h := newHarness(t)

	if code := h.run("", "frobnicate"); code == 0 {
		t.Error("unknown subcommand exited 0, want non-zero")
	}
	if !strings.Contains(h.stderr.String(), "frobnicate") {
		t.Errorf("stderr = %q, want it to name the bad subcommand", h.stderr)
	}
}

func TestNoSubcommandPrintsUsage(t *testing.T) {
	h := newHarness(t)

	h.run("")
	if !strings.Contains(h.stderr.String()+h.stdout.String(), "collect") {
		t.Error("bare invocation should show usage listing the subcommands")
	}
}

func TestVersion(t *testing.T) {
	h := newHarness(t)

	if code := h.run("", "version"); code != 0 {
		t.Fatalf("version exit = %d", code)
	}
	if strings.TrimSpace(h.stdout.String()) == "" {
		t.Error("version printed nothing")
	}
}

func TestCollectRefreshPokesTheBarAfterWritingState(t *testing.T) {
	h := newHarness(t)

	if code := h.run(payloadJSON(63), "collect", "--print", "none", "--refresh", "pkill -SIGRTMIN+12 i3blocks"); code != 0 {
		t.Fatalf("collect exit = %d, stderr = %s", code, h.stderr)
	}
	if len(h.shell) != 1 || h.shell[0] != "pkill -SIGRTMIN+12 i3blocks" {
		t.Fatalf("shell calls = %v, want the refresh command", h.shell)
	}
	if _, err := os.Stat(filepath.Join(h.dir, "state.json")); err != nil {
		t.Errorf("state was not written before the refresh: %v", err)
	}
}

func TestCollectDoesNotPokeTheBarWhenThePayloadIsRejected(t *testing.T) {
	h := newHarness(t)

	if code := h.run("{not json", "collect", "--refresh", "pkill -SIGRTMIN+12 i3blocks"); code == 0 {
		t.Fatal("collect exited 0 on malformed input")
	}
	if len(h.shell) != 0 {
		t.Errorf("shell calls = %v, want none — nothing changed, so the bar has nothing to re-read", h.shell)
	}
}

func TestCollectSurvivesAFailingRefreshCommand(t *testing.T) {
	h := newHarness(t)
	h.shellErr = errors.New("no such process")

	if code := h.run(payloadJSON(63), "collect", "--print", "none", "--refresh", "pkill -SIGRTMIN+12 i3blocks"); code != 0 {
		t.Errorf("collect exit = %d, want 0 — a dead bar must not break Claude Code's status line", code)
	}
}

func TestCollectPassthroughFeedsTheOriginalPayloadOnward(t *testing.T) {
	h := newHarness(t)
	in := payloadJSON(63)

	if code := h.run(in, "collect", "--passthrough", "my-statusline.sh"); code != 0 {
		t.Fatalf("collect exit = %d, stderr = %s", code, h.stderr)
	}
	if len(h.shell) != 1 || h.shell[0] != "my-statusline.sh" {
		t.Fatalf("shell calls = %v, want the passthrough command", h.shell)
	}
	if !strings.Contains(h.stdout.String(), fmt.Sprintf("passthrough saw %d bytes", len(in))) {
		t.Errorf("stdout = %q, want the passthrough output built from the full payload", h.stdout)
	}
}

func TestDetailNotifyFallbackStillReachesStdout(t *testing.T) {
	h := newHarness(t)
	h.notifier.fallback = true
	h.run(payloadJSON(63), "collect", "--print", "none")

	h.run("", "detail", "--notify")
	if !strings.Contains(h.stdout.String(), "Weekly") {
		t.Errorf("stdout = %q, want the breakdown — on a headless box the fallback is the whole point", h.stdout)
	}
}
