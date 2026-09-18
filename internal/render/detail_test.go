package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

func detail(s usage.Snapshot, o Options) string {
	return NewRenderer(o).Detail(s).String()
}

func TestDetailGolden(t *testing.T) {
	got := detail(live(), opts())
	golden := filepath.Join("testdata", "detail.txt")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Errorf("Detail mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestDetailSaysHowOldTheDataIs(t *testing.T) {
	if !strings.Contains(detail(live(), opts()), "11:58") {
		t.Errorf("Detail should state the observation time, got:\n%s", detail(live(), opts()))
	}
}

func TestDetailCarriesOnlyAccountWideFacts(t *testing.T) {
	got := detail(live(), opts())
	for _, leaked := range []string{"Opus", "2.41", "Context", "45m"} {
		if strings.Contains(got, leaked) {
			t.Errorf("Detail leaked %q — per-session figures belong to whichever session wrote last, which is arbitrary with several open:\n%s", leaked, got)
		}
	}
	for _, want := range []string{"5h window", "Weekly", "as of"} {
		if !strings.Contains(got, want) {
			t.Errorf("Detail is missing %q:\n%s", want, got)
		}
	}
}

func TestDetailShowsTheSpendLimitOnlyWhenThereIsOne(t *testing.T) {
	if strings.Contains(detail(live(), opts()), "Spend") {
		t.Error("Detail mentions Spend with no spend limit in the snapshot")
	}

	s := live()
	s.SpendLimit = usage.NewLimit(62.8, at(500*time.Hour), now)
	if !strings.Contains(detail(s, opts()), "Spend") {
		t.Error("Detail omits the spend limit when the snapshot carries one")
	}
}

func TestDetailLinePadsByRuneNotByte(t *testing.T) {
	d := NewDetail(live(), opts())
	l := usage.NewLimit(63, at(time.Hour), now)

	col := func(line string) int { return len([]rune(line[:strings.Index(line, "%")])) }
	if col(d.line("Weekly", l)) != col(d.line("Sessão", l)) {
		t.Errorf("percentage column differs between an ascii and a multi-byte label — padding must count runes, not bytes\n%s%s",
			d.line("Weekly", l), d.line("Sessão", l))
	}
}
