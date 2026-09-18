package render

import "html"

type Markup string

const (
	MarkupNone  Markup = "none"
	MarkupPango Markup = "pango"
)

func (m Markup) Pango() bool { return m == MarkupPango }

func (m Markup) Escape(s string) string {
	if !m.Pango() {
		return s
	}
	return html.EscapeString(s)
}
