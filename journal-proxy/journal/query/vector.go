package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"journal-proxy/global"
	"journal-proxy/journal/operator"
	"journal-proxy/journal/operator/types"
	"strings"
)

type VectorDirection string

const (
	Before VectorDirection = "BEFORE"
	After  VectorDirection = "AFTER"
)

func (v VectorDirection) Validate() error {
	switch v {
	case Before, After:
		break
	default:
		return fmt.Errorf("invalid vector direction: %s", v)
	}
	return nil
}

type Vector struct {
	Base
	Time      global.FrontUint64
	Direction VectorDirection
}

func (q Vector) Validate() error {
	if err := q.Base.Validate(); err != nil {
		return err
	}
	if err := q.Direction.Validate(); err != nil {
		return err
	}
	if q.Time == 0 {
		return errors.New("time is required")
	}
	return nil
}

func (q Vector) SQL() (string, []any) {
	operator := ">"
	if q.Direction == Before {
		operator = "<"
	}

	sql := []string{
		q.Select(),
		q.From(),
	}

	cond := []string{
		fmt.Sprintf("time %s ?", operator),
	}

	args := []any{q.Time}

	if q.Where != nil {
		if c, a := q.Where.SQL(); c != "" {
			cond = append(cond, c)
			args = append(args, a...)
		}
	}
	sql = append(sql, "WHERE "+strings.Join(cond, " AND "))

	sql = append(sql, q.OrderBy())
	if q.Direction == Before {
		sql = append(sql, "DESC")
	}

	sql = append(sql, q.Limit())
	return strings.Join(sql, " "), args
}

func (q *Vector) UnmarshalJSON(b []byte) error {
	type alias Vector
	var data struct {
		alias
		Where operator.Raw
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	*q = Vector(data.alias)
	err = data.Where.Unmarshal(&q.Where)
	if err != nil {
		return err
	}
	return nil
}

func (q *Vector) Filter() types.FilterFunc {
	return func(entry *global.Entry) bool {
		if q.Where != nil && !q.Where.Filter()(entry) {
			return false
		}
		switch q.Direction {
		case Before:
			return entry.Time < (uint64)(q.Time)
		case After:
			return entry.Time > (uint64)(q.Time)
		}
		return false
	}
}

func (q *Vector) EnvID() string {
	return q.Base.EnvID
}
