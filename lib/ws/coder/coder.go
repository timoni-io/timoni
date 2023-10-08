package coder

import (
	"bytes"
	"errors"
	"io"
)

type Coder interface {
	// Decode decodes value read from r into v
	Decode(r io.Reader, v any) error
	// Encode encodes v value into w
	Encode(w io.Writer, v any) error
}

type Raw struct {
	data   []byte
	reader io.Reader
}

// NewRaw changes []byte into Raw
func NewRaw(p []byte) Raw {
	return Raw{data: p}
}

func (raw Raw) Read(p []byte) (int, error) {
	if raw.reader == nil {
		// runs on every request
		raw.reader = bytes.NewReader(raw.data)
	}
	return raw.reader.Read(p)
}

func (raw *Raw) Write(p []byte) (int, error) {
	raw.data = append(raw.data, p...)
	return len(p), nil
}

func (raw *Raw) Len() int {
	return len(raw.data)
}

func (raw Raw) MarshalJSON() ([]byte, error) {
	return raw.data, nil
}

func (raw *Raw) UnmarshalJSON(data []byte) error {
	if raw == nil {
		return errors.New("coder.Raw: UnmarshalJSON on nil pointer")
	}
	raw.data = append(raw.data[0:0], data...)
	return nil
}

func (raw Raw) MarshalBinary() ([]byte, error) {
	return raw.data, nil
}

func (raw *Raw) UnmarshalBinary(data []byte) error {
	if raw == nil {
		return errors.New("coder.Raw: UnmarshalBinary on nil pointer")
	}
	raw.data = append(raw.data[0:0], data...)
	return nil
}
