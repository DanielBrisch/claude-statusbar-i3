package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/statusline"
	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

const (
	Version          = 1
	SessionRetention = 30 * time.Minute
)

type limit struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
	ObservedAt     int64   `json:"observed_at"`
}

type account struct {
	FiveHour   *limit `json:"five_hour"`
	SevenDay   *limit `json:"seven_day"`
	SpendLimit *limit `json:"spend_limit"`
}

type session struct {
	Model             string  `json:"model"`
	CostUSD           float64 `json:"cost_usd"`
	ContextUsedPct    float64 `json:"context_used_pct"`
	ContextWindowSize int64   `json:"context_window_size"`
	DurationMS        int64   `json:"duration_ms"`
	CWD               string  `json:"cwd"`
	UpdatedAt         int64   `json:"updated_at"`
	Seq               int64   `json:"seq"`
}

type file struct {
	Version   int                `json:"version"`
	UpdatedAt int64              `json:"updated_at"`
	Seq       int64              `json:"seq"`
	Account   account            `json:"account"`
	Sessions  map[string]session `json:"sessions"`
}

type Store struct {
	path string
}

func New(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Path() string { return s.path }

func (s *Store) Merge(p statusline.Payload, now time.Time) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	lock, err := os.OpenFile(s.path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("open state lock: %w", err)
	}
	defer lock.Close()
	if err := lockFile(lock); err != nil {
		return fmt.Errorf("lock state: %w", err)
	}
	defer unlockFile(lock)

	f, err := s.read()
	if err != nil {
		return err
	}

	f.Version = Version
	f.UpdatedAt = now.Unix()
	f.Seq++
	if p.RateLimits != nil {
		f.Account = account{
			FiveHour:   toStored(p.RateLimits.FiveHour, now),
			SevenDay:   toStored(p.RateLimits.SevenDay, now),
			SpendLimit: toStored(p.RateLimits.SpendLimit, now),
		}
	}
	if p.SessionID != "" {
		if f.Sessions == nil {
			f.Sessions = map[string]session{}
		}
		f.Sessions[p.SessionID] = session{
			Model:             p.Model,
			CostUSD:           p.CostUSD,
			ContextUsedPct:    p.ContextUsedPct,
			ContextWindowSize: p.ContextWindowSize,
			DurationMS:        p.DurationMS,
			CWD:               p.CWD,
			UpdatedAt:         now.Unix(),
			Seq:               f.Seq,
		}
	}
	for id, sess := range f.Sessions {
		if now.Sub(time.Unix(sess.UpdatedAt, 0)) > SessionRetention {
			delete(f.Sessions, id)
		}
	}

	return s.write(f)
}

func (s *Store) Snapshot() (usage.Snapshot, error) {
	f, err := s.read()
	if err != nil {
		return usage.Snapshot{}, err
	}
	return usage.Snapshot{
		FiveHour:   toDomain(f.Account.FiveHour),
		SevenDay:   toDomain(f.Account.SevenDay),
		SpendLimit: toDomain(f.Account.SpendLimit),
		Session:    newestSession(f.Sessions),
		UpdatedAt:  time.Unix(f.UpdatedAt, 0),
	}, nil
}

func (s *Store) read() (file, error) {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return file{Version: Version, Sessions: map[string]session{}}, nil
	}
	if err != nil {
		return file{}, fmt.Errorf("read state file %s: %w", s.path, err)
	}
	var f file
	if err := json.Unmarshal(b, &f); err != nil {
		return file{}, fmt.Errorf("parse state file %s: %w", s.path, err)
	}
	if f.Sessions == nil {
		f.Sessions = map[string]session{}
	}
	return f, nil
}

func (s *Store) write(f file) error {
	b, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".state-*.json")
	if err != nil {
		return fmt.Errorf("create temp state: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp state: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp state: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return fmt.Errorf("chmod temp state: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	return nil
}

func toStored(l *statusline.Limit, now time.Time) *limit {
	if l == nil {
		return nil
	}
	return &limit{UsedPercentage: l.UsedPercentage, ResetsAt: l.ResetsAt, ObservedAt: now.Unix()}
}

func toDomain(l *limit) *usage.Limit {
	if l == nil {
		return nil
	}
	return &usage.Limit{
		UsedPercentage: l.UsedPercentage,
		ResetsAt:       time.Unix(l.ResetsAt, 0),
		ObservedAt:     time.Unix(l.ObservedAt, 0),
	}
}

func newestSession(sessions map[string]session) *usage.Session {
	var newest *session
	for _, sess := range sessions {
		if newest == nil || sess.Seq > newest.Seq {
			newest = &sess
		}
	}
	if newest == nil {
		return nil
	}
	return &usage.Session{
		Model:             newest.Model,
		CostUSD:           newest.CostUSD,
		ContextUsedPct:    newest.ContextUsedPct,
		ContextWindowSize: newest.ContextWindowSize,
		Duration:          time.Duration(newest.DurationMS) * time.Millisecond,
		CWD:               newest.CWD,
		UpdatedAt:         time.Unix(newest.UpdatedAt, 0),
	}
}
