package query

import (
	"encoding/json"
	"journal-proxy/journal/operator/types"
)

type Multi struct {
	Queries []QueryI
}

func (q Multi) LimitRow() int {
	return 40
}
func (q Multi) IsEvent() bool {
	return false
}
func (q Multi) Validate() error {
	return nil
}
func (q Multi) SQL() (string, []any) {
	return "", nil
}
func (q Multi) Filter() types.FilterFunc {
	return nil
}
func (q Multi) EnvID() string {
	return ""
}
func (q *Multi) UnmarshalJSON(b []byte) error {
	type alias Multi
	var qs struct {
		alias
		Queries []json.RawMessage
	}

	err := json.Unmarshal(b, &qs)
	if err != nil {
		return err
	}

	q.Queries = make([]QueryI, len(qs.Queries))

	for i, query := range qs.Queries {
		err := Unmarshal(query, &q.Queries[i])
		if err != nil {
			return err
		}
	}

	return nil
}
