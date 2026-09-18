package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/statusline"
)

var now = time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)

func newStore(t *testing.T) *Store {
	t.Helper()
	return New(filepath.Join(t.TempDir(), "state.json"))
}

func payload(sessionID string, fiveHourPct float64) statusline.Payload {
	return statusline.Payload{
		SessionID:      sessionID,
		Model:          "Opus",
		CostUSD:        1.23,
		DurationMS:     45000,
		ContextUsedPct: 8,
		RateLimits: &statusline.RateLimits{
			FiveHour: &statusline.Limit{UsedPercentage: fiveHourPct, ResetsAt: now.Add(90 * time.Minute).Unix()},
			SevenDay: &statusline.Limit{UsedPercentage: 41.2, ResetsAt: now.Add(96 * time.Hour).Unix()},
		},
	}
}

func TestSnapshotOfMissingFileIsEmptyAndNotAnError(t *testing.T) {
	s := newStore(t)

	snap, err := s.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() on a missing file returned error %v, want nil", err)
	}
	if snap.HasVisibleWindow(now) {
		t.Error("Snapshot() of a missing file has a visible window, want none")
	}
}

func TestMergeThenSnapshotRoundTripsRateLimits(t *testing.T) {
	s := newStore(t)

	if err := s.Merge(payload("a", 23.5), now); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	snap, err := s.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if snap.FiveHour == nil {
		t.Fatal("FiveHour is nil after Merge")
	}
	if got, want := snap.FiveHour.UsedPercentage, 23.5; got != want {
		t.Errorf("FiveHour.UsedPercentage = %v, want %v", got, want)
	}
	if got, want := snap.FiveHour.ResetsAt.Unix(), now.Add(90*time.Minute).Unix(); got != want {
		t.Errorf("FiveHour.ResetsAt = %v, want %v", got, want)
	}
	if snap.SpendLimit != nil {
		t.Errorf("SpendLimit = %+v, want nil when the payload omitted it", snap.SpendLimit)
	}
}

func TestMergeWithoutRateLimitsKeepsWhatWasAlreadyKnown(t *testing.T) {
	s := newStore(t)
	if err := s.Merge(payload("a", 23.5), now); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	bare := statusline.Payload{SessionID: "b", Model: "Opus"}
	if err := s.Merge(bare, now.Add(time.Minute)); err != nil {
		t.Fatalf("Merge bare payload: %v", err)
	}

	snap, err := s.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if snap.FiveHour == nil {
		t.Fatal("FiveHour was wiped by a payload that carries no rate_limits")
	}
	if got, want := snap.FiveHour.UsedPercentage, 23.5; got != want {
		t.Errorf("FiveHour.UsedPercentage = %v, want %v", got, want)
	}
}

func TestMergeKeepsSessionsApart(t *testing.T) {
	s := newStore(t)
	if err := s.Merge(payload("a", 10), now); err != nil {
		t.Fatalf("Merge a: %v", err)
	}
	if err := s.Merge(payload("b", 20), now.Add(time.Second)); err != nil {
		t.Fatalf("Merge b: %v", err)
	}

	raw := readFile(t, s)
	if len(raw.Sessions) != 2 {
		t.Fatalf("stored %d sessions, want 2: %+v", len(raw.Sessions), raw.Sessions)
	}
}

func TestSnapshotUsesTheMostRecentlyActiveSession(t *testing.T) {
	s := newStore(t)
	older := payload("a", 10)
	older.CostUSD = 1
	newer := payload("b", 10)
	newer.CostUSD = 9

	if err := s.Merge(older, now); err != nil {
		t.Fatalf("Merge a: %v", err)
	}
	if err := s.Merge(newer, now.Add(time.Minute)); err != nil {
		t.Fatalf("Merge b: %v", err)
	}

	snap, err := s.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if snap.Session == nil {
		t.Fatal("Session is nil")
	}
	if got, want := snap.Session.CostUSD, 9.0; got != want {
		t.Errorf("Session.CostUSD = %v, want %v (the freshest session)", got, want)
	}
}

func TestMergePrunesSessionsIdleBeyondTheRetentionWindow(t *testing.T) {
	s := newStore(t)
	if err := s.Merge(payload("stale", 10), now); err != nil {
		t.Fatalf("Merge stale: %v", err)
	}
	if err := s.Merge(payload("fresh", 10), now.Add(SessionRetention+time.Minute)); err != nil {
		t.Fatalf("Merge fresh: %v", err)
	}

	raw := readFile(t, s)
	if _, ok := raw.Sessions["stale"]; ok {
		t.Error("stale session survived pruning")
	}
	if _, ok := raw.Sessions["fresh"]; !ok {
		t.Error("fresh session was pruned")
	}
}

func TestConcurrentMergesNeverCorruptTheFile(t *testing.T) {
	s := newStore(t)
	const n = 24

	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := s.Merge(payload(string(rune('a'+i)), float64(i)), now.Add(time.Duration(i)*time.Second)); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent Merge: %v", err)
	}

	raw := readFile(t, s)
	if len(raw.Sessions) != n {
		t.Errorf("stored %d sessions, want %d — a concurrent write lost data", len(raw.Sessions), n)
	}
}

func TestSnapshotRejectsUnreadableState(t *testing.T) {
	s := newStore(t)
	if err := os.WriteFile(s.Path(), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("seed corrupt file: %v", err)
	}

	if _, err := s.Snapshot(); err == nil {
		t.Fatal("Snapshot() of a corrupt file returned nil error, want an error")
	} else if !strings.Contains(err.Error(), "state") {
		t.Errorf("error %q should mention the state file", err)
	}
}

func readFile(t *testing.T, s *Store) file {
	t.Helper()
	b, err := os.ReadFile(s.Path())
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var f file
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatalf("state file is not valid JSON: %v\n%s", err, b)
	}
	return f
}

func TestSnapshotBreaksUpdatedAtTiesByWriteOrder(t *testing.T) {
	s := newStore(t)
	first := payload("a", 10)
	first.CostUSD = 1
	second := payload("b", 10)
	second.CostUSD = 9

	if err := s.Merge(first, now); err != nil {
		t.Fatalf("Merge a: %v", err)
	}
	if err := s.Merge(second, now); err != nil {
		t.Fatalf("Merge b: %v", err)
	}

	for i := range 50 {
		snap, err := s.Snapshot()
		if err != nil {
			t.Fatalf("Snapshot: %v", err)
		}
		if snap.Session == nil {
			t.Fatal("Session is nil")
		}
		if got, want := snap.Session.CostUSD, 9.0; got != want {
			t.Fatalf("iteration %d: Session.CostUSD = %v, want %v — sessions written in the same second must resolve by write order, not map iteration order", i, got, want)
		}
	}
}
