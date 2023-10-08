package types

import "journal-proxy/global"

type OperatorT string

const (
	AND     OperatorT = "AND"
	OR      OperatorT = "OR"
	NOT     OperatorT = "NOT"
	IS      OperatorT = "IS"
	EXISTS  OperatorT = "EXISTS"
	BETWEEN OperatorT = "BETWEEN"
)

type FilterFunc func(entry *global.Entry) bool

type Operator interface {
	SQL() (string, []any)
	MarshalJSON() ([]byte, error)
	Filter() FilterFunc
}
