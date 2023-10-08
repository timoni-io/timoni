package conv

import (
	"regexp"
	"strings"
	"unicode"
)

// String converts all characters to lowercase and replaces all non-alphanumeric characters with an underscore
// it also removes all leading and trailing non-alphanumeric characters
func String[T ~string](key T) T {
	k := strings.TrimSpace(string(key))
	if k == "" {
		return ""
	}

	repeats := true

	k = strings.Map(func(chr rune) rune {
		// to lower
		if chr >= 'A' && chr <= 'Z' {
			repeats = false
			return chr ^ 0x20
		}

		// pass a-z, 0-9
		if (chr >= 'a' && chr <= 'z') || (chr >= '0' && chr <= '9') {
			repeats = false
			return chr
		}

		// replace non-alphanumeric with underscore
		if !repeats {
			repeats = true
			return '_'
		}
		return -1
	}, k)

	// remove trailing underscore
	if k[len(k)-1] == '_' {
		return T(k[:len(k)-1])
	}

	return T(k)
}

var nonAlphanumericRegex = regexp.MustCompile(`[^\p{L}\p{N}]+`)
var plReplacer = strings.NewReplacer("ą", "a", "ć", "c", "ę", "e", "ł", "l", "ń", "n", "ó", "o", "ś", "s", "ź", "z", "ż", "z")

func KeyString(s string) string {
	// pozostawia tylko male literki, cyfry i -
	key := strings.ToLower(s)
	key = plReplacer.Replace(key)
	key = nonAlphanumericRegex.ReplaceAllString(key, "-")
	key = strings.Trim(key, "-")
	return key
}

func CleanString(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) || r == '\n' {
			return r
		}
		return -1
	}, str)
}
