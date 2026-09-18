package setup

type statusLine struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func newStatusLine(command string) statusLine {
	return statusLine{Type: "command", Command: command}
}
