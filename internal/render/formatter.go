package render

import "github.com/DanielBrisch/claude-statusbar-i3/internal/usage"

type Formatter interface {
	Render(usage.Snapshot) Output
}
