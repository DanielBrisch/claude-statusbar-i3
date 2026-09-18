package statusline

import (
	"encoding/json"
	"fmt"
	"io"
)

type Parser struct{}

func NewParser() Parser {
	return Parser{}
}

func (Parser) Parse(r io.Reader) (Payload, error) {
	var w wirePayload
	if err := json.NewDecoder(r).Decode(&w); err != nil {
		return Payload{}, fmt.Errorf("decode statusline payload: %w", err)
	}
	return w.payload(), nil
}
