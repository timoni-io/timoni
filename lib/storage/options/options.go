package options

import "time"

type TTLOption struct {
	Value time.Duration
}

func TTL(value time.Duration) *TTLOption {
	return &TTLOption{
		Value: value,
	}
}
