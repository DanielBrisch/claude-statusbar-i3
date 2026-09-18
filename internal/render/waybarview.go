package render

type waybarView struct {
	Text       string  `json:"text"`
	Tooltip    string  `json:"tooltip"`
	Class      string  `json:"class"`
	Percentage float64 `json:"percentage"`
}
