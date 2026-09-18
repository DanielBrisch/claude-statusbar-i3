package notify

import (
	"errors"
	"strings"
	"testing"
)

type call struct {
	name string
	args []string
}

func recorder(available map[string]bool, failing map[string]bool) (*Notifier, *[]call) {
	var calls []call
	n := &Notifier{
		Lookup: func(name string) (string, error) {
			if available[name] {
				return "/usr/bin/" + name, nil
			}
			return "", errors.New("not found")
		},
		Run: func(name string, args ...string) error {
			calls = append(calls, call{name: name, args: args})
			if failing[name] {
				return errors.New("boom")
			}
			return nil
		},
	}
	return n, &calls
}

func TestSendPrefersDunstify(t *testing.T) {
	n, calls := recorder(map[string]bool{"dunstify": true, "notify-send": true}, nil)

	if err := n.Send("Claude usage", "body"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("made %d calls, want 1: %+v", len(*calls), *calls)
	}
	if (*calls)[0].name != "dunstify" {
		t.Errorf("called %q, want dunstify", (*calls)[0].name)
	}
}

func TestSendReusesOneNotificationSlot(t *testing.T) {
	n, calls := recorder(map[string]bool{"dunstify": true}, nil)

	if err := n.Send("Claude usage", "body"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	joined := strings.Join((*calls)[0].args, " ")
	if !strings.Contains(joined, "--replace") {
		t.Errorf("dunstify args %q should carry --replace so repeated clicks do not stack notifications", joined)
	}
}

func TestSendFallsBackToNotifySendWhenDunstifyIsMissing(t *testing.T) {
	n, calls := recorder(map[string]bool{"notify-send": true}, nil)

	if err := n.Send("Claude usage", "body"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(*calls) != 1 || (*calls)[0].name != "notify-send" {
		t.Errorf("calls = %+v, want a single notify-send", *calls)
	}
}

func TestSendFallsThroughWhenTheNotifierFails(t *testing.T) {
	n, calls := recorder(
		map[string]bool{"dunstify": true, "notify-send": true},
		map[string]bool{"dunstify": true},
	)

	if err := n.Send("Claude usage", "body"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(*calls) != 2 {
		t.Fatalf("made %d calls, want 2 (dunstify then notify-send): %+v", len(*calls), *calls)
	}
	if (*calls)[1].name != "notify-send" {
		t.Errorf("second call = %q, want notify-send", (*calls)[1].name)
	}
}

func TestSendPrintsToStdoutWhenNoNotifierExists(t *testing.T) {
	n, calls := recorder(nil, nil)
	var out strings.Builder
	n.Out = &out

	if err := n.Send("Claude usage", "sessão 63%"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(*calls) != 0 {
		t.Errorf("calls = %+v, want none", *calls)
	}
	if !strings.Contains(out.String(), "sessão 63%") {
		t.Errorf("stdout = %q, want the body", out.String())
	}
}
