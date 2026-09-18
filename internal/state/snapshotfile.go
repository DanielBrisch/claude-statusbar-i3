package state

import (
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/statusline"
	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

const Version = 1

type snapshotFile struct {
	Version   int      `json:"version"`
	UpdatedAt int64    `json:"updated_at"`
	Seq       int64    `json:"seq"`
	Account   account  `json:"account"`
	Sessions  sessions `json:"sessions"`
}

func newSnapshotFile() snapshotFile {
	return snapshotFile{Version: Version, Sessions: newSessions()}
}

func (f *snapshotFile) absorb(p statusline.Payload, now time.Time) {
	f.Version = Version
	f.UpdatedAt = now.Unix()
	f.Seq++

	if p.HasRateLimits() {
		f.Account = newAccount(p.RateLimits, now)
	}
	if f.Sessions == nil {
		f.Sessions = newSessions()
	}
	if p.HasSession() {
		f.Sessions.record(p.SessionID, newStoredSession(p, now, f.Seq))
	}
	f.Sessions.prune(now)
}

func (f snapshotFile) domain() usage.Snapshot {
	return usage.NewSnapshot(
		f.Account.FiveHour.domain(),
		f.Account.SevenDay.domain(),
		f.Account.SpendLimit.domain(),
		f.Sessions.newest(),
		time.Unix(f.UpdatedAt, 0),
	)
}
