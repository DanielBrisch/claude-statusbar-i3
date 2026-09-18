package render

import (
	"fmt"
	"strings"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/usage"
)

type Detail struct {
	snapshot usage.Snapshot
	opts     Options
}

func NewDetail(s usage.Snapshot, o Options) Detail {
	return Detail{snapshot: s, opts: o}
}

func (d Detail) String() string {
	var b strings.Builder
	b.WriteString(d.line("5h window", d.snapshot.FiveHour))
	b.WriteString(d.line("Weekly", d.snapshot.SevenDay))
	if d.snapshot.SpendLimit != nil {
		b.WriteString(d.line("Spend", d.snapshot.SpendLimit))
	}
	fmt.Fprintf(&b, "as of %s\n", d.observedAt())
	return b.String()
}

func (d Detail) line(label string, l *usage.Limit) string {
	if l.Expired(d.opts.Now) {
		return fmt.Sprintf("%-11s %5s\n", label, Placeholder)
	}
	return fmt.Sprintf("%-11s %5s  resets %s (%s)\n", label, l.Percent(),
		l.ResetsAt.Format("Jan 02 15:04"), l.TimeLeft(d.opts.Now))
}

func (d Detail) observedAt() string {
	return newStamp(d.snapshot.UpdatedAt).Clock()
}
