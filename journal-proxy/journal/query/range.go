package query

import (
	"encoding/json"
	"errors"
	"journal-proxy/global"
	"journal-proxy/journal/operator"
	"journal-proxy/journal/operator/types"
	"strings"
)

type Range struct {
	Base
	TimeBegin global.FrontUint64 // in ns
	TimeEnd   global.FrontUint64 // in ns
}

func (q Range) Validate() error {
	if err := q.Base.Validate(); err != nil {
		return err
	}
	if q.TimeBegin == 0 && q.TimeEnd == 0 {
		return errors.New("time begin or time end are required")
	}
	return nil
}

func (q Range) SQL() (string, []any) {
	sql := []string{
		q.Select(),
		q.From(),
	}

	var args []any

	cond := []string{}
	if q.TimeBegin > 0 {
		cond = append(cond, "time >= ?")
		args = append(args, q.TimeBegin)
	}
	if q.TimeEnd > 0 {
		cond = append(cond, "time <= ?")
		args = append(args, q.TimeEnd)
	}

	if q.Where != nil {
		if c, a := q.Where.SQL(); c != "" {
			cond = append(cond, c)
			args = append(args, a...)
		}
	}
	sql = append(sql, "WHERE "+strings.Join(cond, " AND "))

	sql = append(sql, q.OrderBy())
	sql = append(sql, q.Limit())
	return strings.Join(sql, " "), args
}

func (q *Range) UnmarshalJSON(b []byte) error {
	type alias Range
	var data struct {
		alias
		Where operator.Raw
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	*q = Range(data.alias)
	err = data.Where.Unmarshal(&q.Where)
	if err != nil {
		return err
	}
	return nil
}

func (q *Range) Filter() types.FilterFunc {
	return func(entry *global.Entry) bool {
		if q.Where != nil && q.Where.Filter()(entry) {
			return (uint64)(q.TimeBegin) <= entry.Time && entry.Time <= (uint64)(q.TimeEnd)
		}
		return false
	}
}

func (q *Range) EnvID() string {
	return q.Base.EnvID
}
