package query

import (
	"encoding/json"
	"errors"
	"journal-proxy/global"
	"journal-proxy/journal/operator"
	"journal-proxy/journal/operator/types"
	"strings"
)

type One struct {
	Base
	Time global.FrontUint64
}

func (q One) Validate() error {
	if err := q.Base.Validate(); err != nil {
		return err
	}
	if q.Time == 0 {
		return errors.New("time is required")
	}
	return nil
}

func (q One) SQL() (string, []any) {
	sql := []string{
		q.Select(),
		q.From(),
		"WHERE time = ?",
		"LIMIT 1",
	}
	return strings.Join(sql, " "), []any{q.Time}
}

func (q *One) UnmarshalJSON(b []byte) error {
	type alias One
	var data struct {
		alias
		Where operator.Raw
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	*q = One(data.alias)
	err = data.Where.Unmarshal(&q.Where)
	if err != nil {
		return err
	}
	return nil
}

func (q *One) Filter() types.FilterFunc {
	return func(entry *global.Entry) bool {
		if q.Where != nil && q.Where.Filter()(entry) {
			return (uint64)(q.Time) == entry.Time
		}
		return false
	}
}

func (q *One) EnvID() string {
	return q.Base.EnvID
}
