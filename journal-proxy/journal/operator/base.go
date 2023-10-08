package operator

import (
	"encoding/json"
	"fmt"
	"journal-proxy/global"
	"journal-proxy/journal/operator/types"
	"lib/utils"
	"strings"
)

type (
	AND []types.Operator
	OR  []types.Operator

	NOT struct{ types.Operator }
)

// key: field name, value: if value is string
var serchableColumns = map[string]bool{
	"time":        false,
	"level":       true,
	"message":     true,
	"element":     true,
	"pod":         true,
	"version":     true,
	"git_repo":    true,
	"user_email":  true,
	"tags_string": true,
	"tags_number": true,
}

func (and AND) SQL() (string, []any) {
	sql := make([]string, len(and))
	var args []any
	for i, op := range and {
		q, a := op.SQL()
		sql[i] = q
		args = append(args, a...)
	}
	return fmt.Sprintf("(%s)", strings.Join(sql, " AND ")), args
}

func (or OR) SQL() (string, []any) {
	sql := make([]string, len(or))
	var args []any
	for i, op := range or {
		q, a := op.SQL()
		sql[i] = q
		args = append(args, a...)
	}
	return fmt.Sprintf("(%s)", strings.Join(sql, " OR ")), args
}

func (not NOT) SQL() (string, []any) {
	q, a := not.Operator.SQL()
	return fmt.Sprintf("NOT (%s)", q), a
}

func (and AND) MarshalJSON() ([]byte, error) {
	return MarshalSlice("AND", and)
}

func (and *AND) UnmarshalJSON(data []byte) error {
	return UnmarshalSlice(data, and)
}

func (or OR) MarshalJSON() ([]byte, error) {
	return MarshalSlice("OR", or)
}

func (or *OR) UnmarshalJSON(data []byte) error {
	return UnmarshalSlice(data, or)
}

func (not NOT) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string
		Value types.Operator
	}{
		Type:  "NOT",
		Value: not.Operator,
	})
}

func (not *NOT) UnmarshalJSON(data []byte) error {
	r := &Raw{}
	err := json.Unmarshal(data, r)
	if err != nil {
		return err
	}

	return r.Unmarshal(&not.Operator)
}

func (and AND) Filter() types.FilterFunc {
	filters := make([]types.FilterFunc, 0, len(and))
	for _, op := range and {
		f := op.Filter()
		if f != nil {
			filters = append(filters, f)
		}
	}

	return func(entry *global.Entry) bool {
		return utils.All(filters, func(f types.FilterFunc) bool {
			return f(entry)
		})
	}
}

func (or OR) Filter() types.FilterFunc {
	filters := make([]types.FilterFunc, 0, len(or))
	for _, op := range or {
		f := op.Filter()
		if f != nil {
			filters = append(filters, f)
		}
	}

	return func(entry *global.Entry) bool {
		return utils.Any(filters, func(f types.FilterFunc) bool {
			return f(entry)
		})
	}
}

func (not NOT) Filter() types.FilterFunc {
	return func(entry *global.Entry) bool {
		return !(not.Operator.Filter())(entry)
	}
}
