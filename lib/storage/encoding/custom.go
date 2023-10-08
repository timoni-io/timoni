package encoding

import "log"

// Custom type Encoder/Decoder wrapper
type CustomCoder struct {
	KeyCoder
	ValueCoder
}

func NewCoder(opts ...CoderOpts) Coder {
	coder := &CustomCoder{}

	for _, opt := range opts {
		opt(coder)
	}

	if coder.KeyCoder == nil {
		log.Fatal("Key coder not set")
	}

	if coder.ValueCoder == nil {
		log.Fatal("Value coder not set")
	}

	return coder
}

func (c CustomCoder) EncodeValue(val any) ([]byte, error) {
	if val, ok := val.(CustomValueEncoder); ok {
		return val.EncodeValue(c.ValueCoder)
	}
	return c.ValueCoder.EncodeValue(val)
}

func (c CustomCoder) DecodeValue(data []byte, val any) error {
	if val, ok := val.(CustomValueDecoder); ok {
		return val.DecodeValue(c.ValueCoder, data)
	}
	return c.ValueCoder.DecodeValue(data, val)
}
