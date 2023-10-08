package journal

import (
	"context"
	"journal-proxy/global"
	"journal-proxy/journal/query"
)

type ConnectionI interface {
	Connect() error
	Connected() bool

	CreateTables(envID string)
	DropTables(envID string)

	// Insert
	InsertOne(*global.Entry) error
	InsertMulti([]*global.Entry)

	GetExistingEnvs() []string

	// Returns global.Entry or TagsU
	ExecQuery(context.Context, query.QueryI) (any, error)
}

type TagU struct {
	Keys    []string  `db:"keys"`
	Strings []string  `db:"strings"`
	Numbers []float64 `db:"numbers"`
}

type TagInitU struct {
	Strings map[string][]string  `db:"strings"`
	Numbers map[string][]float64 `db:"numbers"`
}
