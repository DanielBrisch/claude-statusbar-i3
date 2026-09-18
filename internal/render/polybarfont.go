package render

import "fmt"

type polybarFont int

func (f polybarFont) enabled() bool { return f > 0 }

func (f polybarFont) wrap(s string) string {
	if !f.enabled() || s == "" {
		return s
	}
	return fmt.Sprintf("%%{T%d}%s%%{T-}", int(f), s)
}
