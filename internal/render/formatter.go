package render

import "github.com/DanielBrisch/claude-usage-status-i3/internal/usage"

type Formatter interface {
	Render(usage.Snapshot) Output
}
