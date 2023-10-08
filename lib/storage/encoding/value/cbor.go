package value

import (
	"bytes"
	"fmt"
	"io"
	"lib/storage/encoding"
	log "lib/tlog"

	"github.com/fxamacker/cbor/v2"
)

type cborCoder struct {
	cbor.EncMode
}

func CBORCoder() encoding.ValueCoder {
	opts := cbor.CanonicalEncOptions()
	opts.Time = cbor.TimeUnixDynamic

	enc, err := opts.EncMode()
	if err != nil {
		log.Fatal(err)
	}

	return &cborCoder{
		EncMode: enc,
	}
}

// Option for CustomCoder
func CBOR(c *encoding.CustomCoder) {
	c.ValueCoder = CBORCoder()
}

func (c cborCoder) EncodeValue(val any) ([]byte, error) {
	return c.EncMode.Marshal(val)
}

func (c cborCoder) DecodeValue(data []byte, val any) error {
	if len(data) == 0 {
		return nil
	}

	if err := cbor.Valid(data); err != nil {
		return fmt.Errorf("invalid data: %w", err)
	}

	return cbor.Unmarshal(data, val)
}

func (c cborCoder) EncodeBody(val any) (io.Reader, error) {
	buf := &bytes.Buffer{}
	err := c.EncMode.NewEncoder(buf).Encode(val)
	return buf, err
}

func (c cborCoder) DecodeBody(r io.Reader, val any) error {
	return cbor.NewDecoder(r).Decode(val)
}
