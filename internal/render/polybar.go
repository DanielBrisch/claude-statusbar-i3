package render

import "github.com/DanielBrisch/claude-statusbar-i3/internal/usage"

type Polybar struct {
	renderer *Renderer
}

func NewPolybar(r *Renderer) *Polybar {
	o := r.Options().Plain()
	o.Icon = polybarFont(o.IconFont).wrap(o.Icon)
	return &Polybar{renderer: NewRenderer(o)}
}

func (f *Polybar) Render(s usage.Snapshot) Output {
	b := f.renderer.Block(s)
	if b.Empty() {
		return Output{}
	}

	text := b.FullText
	if b.Color != "" {
		text = "%{F" + b.Color + "}" + text + "%{F-}"
	}
	return NewOutput("%{A1:"+DetailCommand+":}"+text+"%{A}", 0)
}
