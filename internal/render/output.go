package render

type Output struct {
	Text     string
	ExitCode int
}

func NewOutput(text string, exitCode int) Output {
	return Output{Text: text, ExitCode: exitCode}
}

func (o Output) Empty() bool { return o.Text == "" }
