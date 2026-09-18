package render

import "github.com/DanielBrisch/claude-usage-status-i3/internal/usage"

type Plain struct {
	renderer *Renderer
}

func NewPlain(r *Renderer) *Plain {
	return &Plain{renderer: NewRenderer(r.Options().Plain())}
}

func (f *Plain) Render(s usage.Snapshot) Output {
	return NewOutput(f.renderer.Block(s).FullText, 0)
}
