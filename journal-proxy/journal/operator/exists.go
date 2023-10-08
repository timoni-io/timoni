package operator

import (
	"encoding/json"
	"fmt"
	"journal-proxy/global"
	"journal-proxy/journal/operator/types"
	"reflect"
	"strings"
)

type EXISTS struct {
	Field string
}

func (exists EXISTS) SQL() (string, []any) {
	isStr, ok := serchableColumns[exists.Field]

	switch {
	case !ok:
		return "(mapContains(tags_string, ?) = 1 OR mapContains(tags_number, ?) = 1)", []any{exists.Field, exists.Field}
	case isStr:
		return fmt.Sprintf("%s != ''", exists.Field), nil
	default:
		return fmt.Sprintf("%s != 0", exists.Field), nil
	}
}

func (exists EXISTS) Filter() types.FilterFunc {
	return func(entry *global.Entry) bool {
		v := reflect.ValueOf(entry).Elem()
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			if strings.EqualFold(t.Field(i).Name, exists.Field) {
				return v.Field(i).IsZero()
			}
		}

		_, ok := entry.TagsString[exists.Field]
		if ok {
			return true
		}
		_, ok = entry.TagsNumber[exists.Field]
		return ok
	}
}

func (exists EXISTS) MarshalJSON() ([]byte, error) {
	type alias EXISTS
	return json.Marshal(struct {
		Type  string
		Value alias
	}{
		Type:  "EXISTS",
		Value: alias(exists),
	})
}
