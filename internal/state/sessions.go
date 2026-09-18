package state

import (
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

const SessionRetention = 30 * time.Minute

type sessions map[string]storedSession

func newSessions() sessions {
	return sessions{}
}

func (s sessions) record(id string, sess storedSession) {
	s[id] = sess
}

func (s sessions) prune(now time.Time) {
	for id, sess := range s {
		if sess.idleFor(now) > SessionRetention {
			delete(s, id)
		}
	}
}

func (s sessions) newest() *usage.Session {
	var newest *storedSession
	for _, sess := range s {
		if newest == nil || sess.Seq > newest.Seq {
			newest = &sess
		}
	}
	if newest == nil {
		return nil
	}
	return newest.domain()
}
