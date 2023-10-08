package value

import (
	"bytes"
	"io"
	"lib/storage/encoding"

	"github.com/pelletier/go-toml/v2"
)

type tomlCoder struct{}

func TOMLCoder() encoding.ValueCoder {
	return &tomlCoder{}
}

// Option for CustomCoder
func TOML(c *encoding.CustomCoder) {
	c.ValueCoder = TOMLCoder()
}

func (c tomlCoder) EncodeValue(val any) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := toml.NewEncoder(buf).Encode(val)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c tomlCoder) DecodeValue(data []byte, val any) error {
	if len(data) == 0 {
		return nil
	}
	return toml.Unmarshal(data, val)
}

func (c tomlCoder) EncodeBody(val any) (io.Reader, error) {
	buf := &bytes.Buffer{}
	err := toml.NewEncoder(buf).Encode(val)
	return buf, err
}

func (c tomlCoder) DecodeBody(r io.Reader, val any) error {
	return toml.NewDecoder(r).Decode(val)
}
