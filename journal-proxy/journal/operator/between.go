package operator

import (
	"encoding/json"
	"fmt"
	"journal-proxy/global"
	"journal-proxy/journal/operator/types"
	"strings"

	"golang.org/x/exp/constraints"
)

type BETWEEN[T constraints.Integer | constraints.Float] struct {
	Field string
	From  *T
	To    *T
}

func (between BETWEEN[T]) SQL() (string, []any) {
	var sql []string

	if between.From == nil && between.To == nil {
		return "", nil
	}

	var args []any

	if between.From != nil {
		sql = append(sql, "? <= ?")
		args = append(args, between.Field, fmt.Sprint(*between.From))
	}
	if between.To != nil {
		sql = append(sql, "? >= ?")
		args = append(args, between.Field, fmt.Sprint(*between.To))
	}

	return fmt.Sprintf("(%s)", strings.Join(sql, " AND ")), args
}

func (between BETWEEN[T]) Filter() types.FilterFunc {
	if between.From == nil && between.To == nil {
		return nil
	}

	return func(entry *global.Entry) bool {
		if strings.ToLower(between.Field) == "time" {
			return entry.Time >= uint64(*between.From) && uint64(*between.To) >= entry.Time
		}

		return entry.TagsNumber[between.Field] >= float64(*between.From) && float64(*between.To) >= entry.TagsNumber[between.Field]
	}
}

func (between BETWEEN[T]) MarshalJSON() ([]byte, error) {
	type alias BETWEEN[T]
	return json.Marshal(struct {
		Type  string
		Value alias
	}{
		Type:  "BETWEEN",
		Value: alias(between),
	})
}
