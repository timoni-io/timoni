package conv

import "strconv"

// Ptr returns pointer to value
func Ptr[T any](v T) *T {
	return &v
}

func Elem[T any](e *T) T {
	if e == nil {
		return *new(T)
	}
	return *e
}

func ToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 0, 64)
	return i
}

func ToBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

// Returns float without trailing zeroes.
//
//	20.60 -> 20.6
//	221.00 -> 221
func Float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
