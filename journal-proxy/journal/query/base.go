package query

import (
	"errors"
	"fmt"
	"journal-proxy/global"
	"journal-proxy/journal/operator/types"
	"lib/utils/conv"
	"lib/utils/math"
)

type Base struct {
	EnvID  string
	Events bool

	FullLog   bool
	LimitRows int `json:"Limit"`

	Where types.Operator
}

func (q Base) LimitRow() int {
	return q.LimitRows
}
func (q Base) IsEvent() bool {
	return q.Events
}

func (q Base) Select() string {
	if q.FullLog {
		return "SELECT *"
	}
	return "SELECT time, level, left(message, 500) as message, element"
}

func (q Base) From() string {
	tablePrefix := "logs"
	if q.Events {
		tablePrefix = "events"
	}
	return fmt.Sprintf("FROM logs.%s_%s", tablePrefix, conv.String(q.EnvID))
}

func (q Base) OrderBy() string {
	return "ORDER BY time"
}

func (q Base) Limit() string {
	return fmt.Sprintf("LIMIT %d", math.Clamp(q.LimitRows, 1, global.JounalProxyConfig.MaxEntriesLimit))
}

func (q Base) Validate() error {
	if q.EnvID == "" {
		return errors.New("missing EnvID")
	}
	return nil
}
