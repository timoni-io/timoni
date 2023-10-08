package query

import (
	"encoding/json"
	"fmt"
	"journal-proxy/journal/operator/types"
)

type QueryT string

const (
	MultiT     QueryT = "MULTI"
	OneT       QueryT = "ONE"
	VectorT    QueryT = "VECTOR"
	RangeT     QueryT = "RANGE"
	TagsT      QueryT = "TAGS"
)

type QueryType struct {
	Type QueryT
}

type QueryI interface {
	Validate() error
	SQL() (string, []any)
	UnmarshalJSON([]byte) error
	LimitRow() int
	IsEvent() bool
	Filter() types.FilterFunc
	EnvID() string
}

func Unmarshal(data []byte, q *QueryI) error {
	var qType QueryType
	err := json.Unmarshal(data, &qType)
	if err != nil {
		return fmt.Errorf("error unmarshalling query type %s: %w", qType, err)
	}

	switch qType.Type {
	case MultiT:
		*q = &Multi{}
	case OneT:
		*q = &One{}
	case VectorT:
		*q = &Vector{}
	case RangeT:
		*q = &Range{}
	case TagsT:
		*q = &Tags{}
	default:
		return fmt.Errorf("incorrect query type %s", qType)
	}

	err = json.Unmarshal(data, q)
	if err != nil {
		return fmt.Errorf("error unmarshalling query: %w", err)
	}

	err = (*q).Validate()
	if err != nil {
		return fmt.Errorf("error validating query: %w", err)
	}
	return nil
}
