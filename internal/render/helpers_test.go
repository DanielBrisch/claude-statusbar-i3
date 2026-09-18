package render

import (
	"time"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

var now = time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)

func at(d time.Duration) time.Time { return now.Add(d) }

func opts() Options { return NewOptions(now) }

func coloured(o Options) Options {
	o.Colors = NewColors("#E5C07B", "#E06C75")
	return o
}

func live() usage.Snapshot {
	return usage.NewSnapshot(
		usage.NewLimit(63, at(102*time.Minute), at(-2*time.Minute)),
		usage.NewLimit(21, at(100*time.Hour), at(-2*time.Minute)),
		nil,
		usage.NewSession("Opus", 2.41, 18, 200000, 45*time.Minute, "", at(-2*time.Minute)),
		at(-2*time.Minute),
	)
}

func block(s usage.Snapshot, o Options) Block { return NewRenderer(o).Block(s) }

func render(t string, s usage.Snapshot, o Options) Output {
	f, err := NewRegistry(o).Lookup(t)
	if err != nil {
		panic(err)
	}
	return f.Render(s)
}
