package value

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"lib/storage/encoding"
)

type jsonCoder struct {
}

func JSONCoder() encoding.ValueCoder {
	return &jsonCoder{}
}

// Option for CustomCoder
func JSON(c *encoding.CustomCoder) {
	c.ValueCoder = JSONCoder()
}

func (c jsonCoder) EncodeValue(val any) ([]byte, error) {
	return json.Marshal(val)
}

func (c jsonCoder) DecodeValue(data []byte, val any) error {
	if len(data) == 0 {
		return nil
	}

	err := json.Unmarshal(data, val)
	if err != nil {
		if !json.Valid(data) {
			return fmt.Errorf("invalid data: %w", err)
		}
	}
	return err
}

func (c jsonCoder) EncodeBody(val any) (io.Reader, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(val)
	return buf, err
}

func (c jsonCoder) DecodeBody(r io.Reader, val any) error {
	return json.NewDecoder(r).Decode(val)
}
