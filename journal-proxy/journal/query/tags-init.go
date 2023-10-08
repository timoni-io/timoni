package query

import (
	"errors"
	"fmt"
	"journal-proxy/global"
	"journal-proxy/journal/operator/types"
	"lib/utils/conv"
	"strings"
)

type TagsInit struct {
	Base
}

func (q TagsInit) Validate() error {
	if err := q.Base.Validate(); err != nil {
		return err
	}
	return nil
}

func (q TagsInit) SQL() (string, []any) {
	sql := []string{
		"SELECT",
		"CAST((mapKeys(s), arrayMap(x->flatten(x), mapValues(s))), 'Map(String, Array(String))') AS strings,",
		"CAST((mapKeys(n), arrayMap(x->flatten(x), mapValues(n))), 'Map(String, Array(Float64))') AS numbers",
		"FROM (",
		"SELECT",
		"groupArraySampleMap(%[1]d)(s) AS s,",
		"groupArraySampleMap(%[1]d)(n) AS n",
		"FROM (",
		"SELECT",
		"groupUniqArrayMap(tags_string) AS s,",
		"groupUniqArrayMap(tags_number) AS n",
		"FROM logs.logs_%[2]s",
		")",
		")",
	}
	return fmt.Sprintf(strings.Join(sql, " "), global.JounalProxyConfig.CacheValuesLimit, conv.String(q.Base.EnvID)), nil
}

func (q *TagsInit) UnmarshalJSON(b []byte) error {
	return errors.New("not implemented")
}

func (q *TagsInit) Filter() types.FilterFunc {
	return nil
}

func (q TagsInit) EnvID() string {
	return q.Base.EnvID
}
