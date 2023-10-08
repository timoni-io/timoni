package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"journal-proxy/global"
	"journal-proxy/journal"
	"journal-proxy/journal/query"
	"lib/tlog"
	"sort"
	"strings"
)

func (c *ConnectionS) ExecQuery(ctx context.Context, q query.QueryI) (any, error) {
	var err error
	var entries any

	tlog.PrintJSON(q)

	switch _q := q.(type) {
	case *query.Multi:
		var entr []any
		for _, query := range _q.Queries {
			e, err := c.ExecQuery(ctx, query)
			if err != nil {
				return entr, err
			}
			entr = append(entr, e)
		}
		return entr, nil

	case *query.Tags:
		var data []*journal.TagU
		data, err = ExecQuery[journal.TagU](ctx, c, q)
		if len(data) <= 0 {
			return nil, fmt.Errorf("no data")
		}
		entries = data[0]
	case *query.One:
		var data []*global.Entry
		data, err = ExecQuery[global.Entry](ctx, c, q)
		if len(data) <= 0 {
			return nil, fmt.Errorf("no data")
		}
		entries = data[0]
	case *query.Vector:
		var data []*global.Entry
		data, err = ExecQuery[global.Entry](ctx, c, q)
		if _q.Direction == query.Before {
			sort.Slice(data, func(i, j int) bool { return data[i].Time < data[j].Time })
		}
		entries = data
	case *query.TagsInit:
		var data []*journal.TagInitU
		data, err = ExecQuery[journal.TagInitU](ctx, c, q)
		if len(data) <= 0 {
			return nil, fmt.Errorf("no data")
		}
		entries = data[0]
	default:
		entries, err = ExecQuery[global.Entry](ctx, c, q)
	}

	if err != nil {
		return nil, err
	}

	return entries, nil
}

type ExecQueryType interface {
	journal.TagU | global.Entry | journal.TagInitU
}

func ExecQuery[T ExecQueryType](ctx context.Context, c *ConnectionS, q query.QueryI) ([]*T, error) {
	if !c.Connected() {
		return nil, errors.New("no db connection")
	}

	quer, args := q.SQL()
	if quer == "" {
		return nil, nil
	}

	tlog.Info(quer)
	fmt.Println(quer, args)

	rows, err := c.DB.Query(ctx, quer, args...)
	if err != nil {
		// silence TABLE DOES NOT EXIST on new envs
		if strings.Contains(err.Error(), "code: 60") {
			return []*T{}, nil
		}
		return nil, fmt.Errorf("query failed %w, %s", err, quer)
	}
	defer rows.Close()

	var entries []*T

	for rows.Next() {

		entry := new(T)
		err := rows.ScanStruct(entry)
		if err != nil {
			// TODO: fix tag cache or rewrite cache system
			tlog.Warning("StructScan failed " + err.Error())
			continue
		}

		switch entr := any(entry).(type) {
		case *global.Entry:
			if entr.TagsString["event"] == "true" {
				entr.Event = true
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
