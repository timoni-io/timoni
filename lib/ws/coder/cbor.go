package coder

import (
	"io"

	"github.com/fxamacker/cbor/v2"
)

type CBOR struct{}

func (CBOR) Decode(r io.Reader, v any) error {
	return cbor.NewDecoder(r).Decode(v)
}

func (CBOR) Encode(w io.Writer, v any) error {
	return cbor.NewEncoder(w).Encode(v)
}
