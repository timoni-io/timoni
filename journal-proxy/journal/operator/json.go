package operator

import (
	"encoding/json"
	"fmt"
	"journal-proxy/journal/operator/types"
	"regexp"
)

// Raw Operator
type Raw struct {
	Type  types.OperatorT
	Value json.RawMessage
}

var betweenFloatRegex = regexp.MustCompile(`\"(?:From|To)\"\:\s{0,}\d{1,}\.\d{0,}`)

func (r Raw) Unmarshal(op *types.Operator) error {
	switch r.Type {
	case types.AND:
		*op = &AND{}
	case types.OR:
		*op = &OR{}
	case types.NOT:
		*op = &NOT{}
	case types.IS:
		*op = &IS{}
	case types.EXISTS:
		*op = &EXISTS{}
	case types.BETWEEN:
		if betweenFloatRegex.Match(r.Value) {
			*op = &BETWEEN[float64]{}
		} else {
			*op = &BETWEEN[int64]{}
		}
	default:
		// Return nil when where is empty
		if len(r.Value) == 0 {
			return nil
		}
		return fmt.Errorf("unknown operator type: %s", r)
	}

	err := json.Unmarshal(r.Value, op)
	if err != nil {
		return err
	}

	return nil
}

type OperatorSlice interface {
	AND | OR
}

func MarshalSlice[T OperatorSlice](t string, op T) ([]byte, error) {
	return json.Marshal(struct {
		Type  string
		Value []types.Operator
	}{
		Type:  t,
		Value: op,
	})

}

func UnmarshalSlice[T OperatorSlice](data []byte, slice *T) error {
	var r []Raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}

	for _, raw := range r {
		var op types.Operator
		err := raw.Unmarshal(&op)
		if err != nil {
			return err
		}
		*slice = append(*slice, op)
	}
	return nil
}
