package conv

import (
	"fmt"
	"time"
)

// UnixTimeStamp ...
func UnixTimeStamp(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}

	return t.Unix()
}

func FixMap(m map[any]any) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		if v, ok := any(v).(map[any]any); ok {
			out[fmt.Sprint(k)] = FixMap(v)
			continue
		}
		out[fmt.Sprint(k)] = v
	}
	return out
}
