package query

import (
	"encoding/json"
	"fmt"
	"journal-proxy/journal/operator"
	"journal-proxy/journal/operator/types"
	"strings"
)

type Tags struct {
	Base
	Field string
}

func (q Tags) Validate() error {
	if err := q.Base.Validate(); err != nil {
		return err
	}
	return nil
}

func (q Tags) SQL() (string, []any) {
	sql := []string{
		q.Select(),
		q.From(),
	}

	var args []any
	if q.Where != nil {
		if c, a := q.Where.SQL(); c != "" {
			sql = append(sql, fmt.Sprintf("WHERE %s", c))
			args = append(args, a...)
		}
	}

	sql = append(sql, "LIMIT 1")
	return strings.Join(sql, " "), args
}

func (q Tags) Select() string {
	selSQL := "groupArray(distinct arrayJoin(arrayConcat(tags_string.keys, tags_number.keys))) as keys"
	if q.Field != "" {
		selSQL = fmt.Sprintf("arrayDistinct(groupArray(tags_string['%[1]s'])) as strings, arrayDistinct(groupArray(tags_number['%[1]s'])) as numbers", q.Field)
	}

	return fmt.Sprintf("SELECT %s", selSQL)
}

func (q *Tags) UnmarshalJSON(b []byte) error {
	type alias Tags
	var data struct {
		alias
		Where operator.Raw
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	*q = Tags(data.alias)
	err = data.Where.Unmarshal(&q.Where)
	if err != nil {
		return err
	}
	return nil
}

func (q *Tags) Filter() types.FilterFunc {
	// TODO: Finish this
	if q.Where == nil {
		return nil
	}
	return q.Where.Filter()
}

func (q *Tags) EnvID() string {
	return q.Base.EnvID
}
