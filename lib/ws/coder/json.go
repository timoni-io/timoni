package coder

import (
	"bufio"
	"encoding/json"
	"io"
)

type JSON struct{}

func (JSON) Decode(r io.Reader, v any) error {
	return json.NewDecoder(bufio.NewReader(r)).Decode(v)
}

func (JSON) Encode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}
