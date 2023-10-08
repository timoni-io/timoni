package operator

import (
	"encoding/json"
	"fmt"
	"journal-proxy/global"
	"journal-proxy/journal/operator/types"
	"lib/utils/conv"
	"lib/utils/math"
	"reflect"
	"strings"
)

type IS struct {
	Field string
	Value string
}

func (is IS) SQL() (string, []any) {
	if _, ok := serchableColumns[is.Field]; ok {
		return fmt.Sprintf("%s = ?", is.Field), []any{is.Value}
	}
	if math.IsNumeric(is.Value) {
		return "tags_number[?] = ?", []any{is.Field, is.Value}
	}
	return "tags_string[?] = ?", []any{is.Field, is.Value}
}

func (is IS) Filter() types.FilterFunc {
	return func(entry *global.Entry) bool {
		v := reflect.ValueOf(entry).Elem()
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if strings.EqualFold(field.Name, is.Field) {
				switch typ := v.Field(i).Interface().(type) {
				case string:
					return typ == is.Value
				case uint64:
					return fmt.Sprint(typ) == is.Value
				}
			}
		}

		if entry.TagsString[is.Field] == is.Value {
			return true
		}

		if conv.Float64ToString(entry.TagsNumber[is.Field]) == is.Value {
			return true
		}

		return false
	}
}

func (is IS) MarshalJSON() ([]byte, error) {
	type alias IS
	return json.Marshal(struct {
		Type  string
		Value alias
	}{
		Type:  "IS",
		Value: alias(is),
	})
}
