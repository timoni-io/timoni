package coder

import (
	"encoding/gob"
	"io"
)

type GOB struct{}

func (GOB) Decode(r io.Reader, v any) error {
	return gob.NewDecoder(r).Decode(v)
}

func (GOB) Encode(w io.Writer, v any) error {
	return gob.NewEncoder(w).Encode(v)
}
