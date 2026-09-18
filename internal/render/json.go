package render

import (
	"encoding/json"

	"github.com/DanielBrisch/claude-statusbar-i3/internal/usage"
)

type JSON struct {
	renderer *Renderer
}

func NewJSON(r *Renderer) *JSON {
	return &JSON{renderer: NewRenderer(r.Options().Plain())}
}

func (f *JSON) Render(s usage.Snapshot) Output {
	o := f.renderer.Options()
	b := f.renderer.Block(s)

	encoded, err := json.MarshalIndent(snapshotView{
		Text:       b.FullText,
		Level:      b.Level.String(),
		Percentage: b.Percentage,
		FiveHour:   newWindowView(s.FiveHour, o.Now),
		SevenDay:   newWindowView(s.SevenDay, o.Now),
		SpendLimit: newWindowView(s.SpendLimit, o.Now),
		Session:    newSessionView(s.Session),
		UpdatedAt:  newStamp(s.UpdatedAt).String(),
	}, "", "  ")
	if err != nil {
		return NewOutput("{}", 0)
	}
	return NewOutput(string(encoded), 0)
}
