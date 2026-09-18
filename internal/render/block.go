package render

import "github.com/DanielBrisch/claude-usage-status-i3/internal/usage"

type Block struct {
	FullText   string
	ShortText  string
	Color      string
	Level      usage.Level
	Percentage float64
}

func (b Block) Empty() bool { return b.FullText == "" }

func (b Block) Urgent() bool { return b.Level == usage.LevelUrgent }
