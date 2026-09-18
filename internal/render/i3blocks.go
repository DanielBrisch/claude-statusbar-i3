package render

import "github.com/DanielBrisch/claude-usage-status-i3/internal/usage"

const i3blocksUrgentExitCode = 33

type I3blocks struct {
	renderer *Renderer
}

func NewI3blocks(r *Renderer) *I3blocks {
	return &I3blocks{renderer: r}
}

func (f *I3blocks) Render(s usage.Snapshot) Output {
	b := f.renderer.Block(s)
	if b.Empty() {
		return Output{}
	}

	text := b.FullText + "\n" + b.ShortText + "\n"
	if b.Color != "" {
		text += b.Color + "\n"
	}

	exit := 0
	if f.renderer.Options().UrgentExit && b.Urgent() {
		exit = i3blocksUrgentExitCode
	}
	return NewOutput(text, exit)
}
