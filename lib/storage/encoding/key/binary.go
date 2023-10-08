package key

import (
	"fmt"
	"lib/storage/encoding"
	"strings"
)

type binary struct {
	encoding.Constants
}

func BinaryCoder() encoding.KeyCoder {
	return &binary{
		encoding.Constants{
			Delimiter:      "\x1C",
			TransactionKey: "TX\x1C",
		},
	}
}

func Binary(c *encoding.CustomCoder) {
	c.KeyCoder = BinaryCoder()
}

func (c binary) Symbols() encoding.Constants {
	return c.Constants
}

func (c binary) EncodeKey(key ...string) string {
	if len(key) == 0 {
		return ""
	}
	keys := make([]string, 0, len(key))
	for _, k := range key {
		if k != "" {
			keys = append(keys, k)
		}
	}

	return strings.Join(keys, c.Delimiter)
}

func (c binary) DecodeKey(key string) []string {
	keys := strings.Split(key, c.Delimiter)
	return keys
}

func (c binary) EncodeBucket(key ...string) string {
	return fmt.Sprintf("[%s]", strings.Join(key, c.Delimiter))
}

func (c binary) DecodeBucket(key ...string) []string {
	keys := []string{}

	for _, k := range key {
		k = strings.Trim(k, "[]")
		for _, b := range strings.Split(k, c.Delimiter) {
			b = strings.Trim(b, "[]")
			if b != "" {
				keys = append(keys, b)
			}
		}
	}

	return keys
}
