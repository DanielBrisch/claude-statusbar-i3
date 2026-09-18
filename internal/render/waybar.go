package render

import (
	"encoding/json"

	"github.com/DanielBrisch/claude-usage-status-i3/internal/usage"
)

type Waybar struct {
	renderer *Renderer
}

func NewWaybar(r *Renderer) *Waybar {
	return &Waybar{renderer: r}
}

func (f *Waybar) Render(s usage.Snapshot) Output {
	b := f.renderer.Block(s)
	if b.Empty() {
		return NewOutput("{}", 0)
	}

	encoded, err := json.Marshal(waybarView{
		Text:       b.FullText,
		Tooltip:    f.renderer.Detail(s).String(),
		Class:      b.Level.String(),
		Percentage: b.Percentage,
	})
	if err != nil {
		return NewOutput("{}", 0)
	}
	return NewOutput(string(encoded), 0)
}
