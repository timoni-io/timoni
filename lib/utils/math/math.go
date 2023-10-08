package math

import (
	"math"

	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

// IsNumeric quickly checks if string is a number. 12x faster then strconv.ParseFloat(s, 64)
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}

	dotFound := false
	for _, v := range s {
		if v == '.' {
			if dotFound {
				return false
			}
			dotFound = true
		} else if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

// RoundTo rounds `n` float to `decimals` number after comma
//
//	RoundTo(1.123, 1) = 1.1
//	RoundTo(1.655, 2) = 1.66
func RoundTo(n float64, decimals uint8) float64 {
	return math.Round(n*math.Pow(10, float64(decimals))) / math.Pow(10, float64(decimals))
}

func Clamp[T Number](v, min, max T) T {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
