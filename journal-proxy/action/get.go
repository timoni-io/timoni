package action

import (
	"encoding/json"
	"journal-proxy/cache"
	"journal-proxy/global"
	"journal-proxy/journal"
	"journal-proxy/journal/query"
	"journal-proxy/metrics"
	"journal-proxy/wsb"
	"lib/terrors"
	log "lib/tlog"
	"lib/utils/conv"
	"lib/utils/slice"
	ws "lib/ws/server"
	"sort"
	"sync"
	"time"
)

type GetSingle struct {
	Query query.QueryI
}

type Get struct {
	Queries []GetSingle `json:"Querys"`
}

func (a Get) Handle(r *ws.Request) (code terrors.Error, data any) {
	conn := wsb.ConnPool.GetNoWait()
	if conn != nil {
		defer wsb.ConnPool.Add(conn)
		return a.queryDB(r, conn)
	}

	return a.queryCache(r)
}

func (a *Get) queryCache(r *ws.Request) (code terrors.Error, data any) {

	limit := 0
	wg := &sync.WaitGroup{}
	lgs := slice.NewSlice[frontEntry](0).Safe()

	for _, get := range a.Queries {
		wg.Add(1)

		go func(q GetSingle) {
			defer wg.Done()
			limit = q.Query.LimitRow()
			filter := q.Query.Filter()
			if filter == nil {
				filter = func(*global.Entry) bool {
					return true
				}
			}

			entries := cache.Cache.Entries(conv.String(q.Query.EnvID()))
			for i := len(entries) - 1; i >= 0 && lgs.Len() < limit; i-- {
				entry := entries[i]
				if filter(entry) {
					lgs.Add(frontEntry{
						Entry: *entry,
						Time:  global.FrontUint64(entry.Time),
					})

				}
			}

		}(get)
	}

	wg.Wait()

	result := lgs.GetAll()
	sort.Slice(result, func(i, j int) bool { return result[i].Time < result[j].Time })

	// cut result to limit
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return terrors.Success, result
}

func (a *Get) queryDB(r *ws.Request, conn journal.ConnectionI) (code terrors.Error, data any) {
	metrics.Vars.SelectRequests.Add(1)

	// FIXME: this is a hack
	lgs := slice.NewSlice[frontEntry](0).Safe()
	limit := 0
	
	ts := time.Now()
	wg := &sync.WaitGroup{}
	for _, get := range a.Queries {
		wg.Add(1)
		go func(g GetSingle) {
			defer wg.Done()
			limit = g.Query.LimitRow()
			data, err := conn.ExecQuery(r.Ctx, g.Query)
			if err != nil {
				log.Error(err)
				// return ws.InternalServerError(r, err)
			}

			switch data := data.(type) {
			case []*global.Entry:
				for _, v := range data {
					v.Event = g.Query.IsEvent()
					lgs.Add(frontEntry{
						Entry: *v,
						Time:  global.FrontUint64(v.Time),
					})
				}

			case *global.Entry:
				lgs.Add(frontEntry{
					Entry: *data,
					Time:  global.FrontUint64(data.Time),
				})

			default:
				log.Error("wrong data type")
			}
		}(get)
	}
	wg.Wait()

	result := lgs.GetAll()
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time > result[j].Time
	})

	metrics.Vars.SelectRequestTime.Add(int64(time.Since(ts).Milliseconds()))

	// cut result to limit
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return terrors.Success, result
}

func (a *Get) UnmarshalJSON(data []byte) error {
	type alias Get
	var action struct {
		alias
		Querys []json.RawMessage
	}

	err := json.Unmarshal(data, &action)
	if err != nil {
		return err
	}

	// Decode every query
	type singleAlias struct {
		GetSingle
		Query json.RawMessage
	}

	for _, get := range action.Querys {
		var q singleAlias
		err = json.Unmarshal(get, &q)
		if err != nil {
			return err
		}
		err := query.Unmarshal(q.Query, &q.GetSingle.Query)
		if err != nil {
			return err
		}
		a.Queries = append(a.Queries, GetSingle{
			Query: q.GetSingle.Query,
		})
	}

	return nil
}
