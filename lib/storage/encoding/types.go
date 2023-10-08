package encoding

import "io"

type Constants struct {
	Delimiter      string
	TransactionKey string
}

type CoderOpts func(*CustomCoder)

// Encoder / Decoder interface
type Coder interface {
	KeyCoder
	ValueCoder
}

type KeyCoder interface {
	// Returns struct with all coder constants
	Symbols() Constants

	EncodeKey(key ...string) string
	DecodeKey(key string) []string

	// Encode/Decode bucket key

	EncodeBucket(key ...string) string
	DecodeBucket(bucket ...string) []string
}

type ValueCoder interface {
	EncodeValue(val any) ([]byte, error)
	DecodeValue(data []byte, val any) error

	EncodeBody(val any) (io.Reader, error)
	DecodeBody(r io.Reader, val any) error
}

type CustomValueEncoder interface {
	EncodeValue(coder ValueCoder) ([]byte, error)
}

type CustomValueDecoder interface {
	DecodeValue(coder ValueCoder, data []byte) error
}
