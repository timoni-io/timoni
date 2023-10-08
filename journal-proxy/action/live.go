package action

import (
	"encoding/json"
	"journal-proxy/cache"
	"journal-proxy/global"
	"journal-proxy/journal/operator"
	"journal-proxy/journal/query"
	"journal-proxy/wsb"
	"lib/terrors"
	log "lib/tlog"
	"lib/utils/conv"
	"lib/utils/math"
	"lib/utils/slice"
	ws "lib/ws/server"
	"sort"
	"sync"
	"time"
)

type Sub struct {
	query.Base
	Elements []string
	LastTime global.FrontUint64 // FIXME: ???
}

type Live struct {
	Subs   []*Sub
	Cancel bool
}

type frontEntry struct {
	global.Entry
	Time global.FrontUint64 `json:"time"`
}

func (a *Live) HandleSub(r *ws.Request, c chan<- *ws.ResponseS) {
	if a.Cancel {
		return
	}
	wg := &sync.WaitGroup{}
	lgs := slice.NewSlice[frontEntry](0).Safe()
	for _, sub := range a.Subs {

		log.PrintJSON(sub)

		wg.Add(1)
		go func(s *Sub) {
			defer wg.Done()
			s.EnvID = conv.String(s.EnvID)
			filter := func(*global.Entry) bool {
				return true
			}

			if s.Where != nil {
				filter = s.Where.Filter()
			}

			if s.Events {
				filter = func(e *global.Entry) bool {
					return e.Event
				}
			}

			for _, entry := range cache.Cache.Entries(s.EnvID, s.Elements...) {
				if filter(entry) {
					lgs.Add(frontEntry{
						Entry: *entry,
						Time:  global.FrontUint64(entry.Time),
					})
				}
			}

			if lgs.Len() < s.LimitRows {
				// Get the rest from db
				conn := wsb.ConnPool.GetNoWait()
				if conn == nil {
					return
				}
				defer wsb.ConnPool.Add(conn)
				data, err := conn.ExecQuery(r.Ctx, s.ToQuery())
				if err != nil {
					log.Error(err)
					c <- &ws.ResponseS{
						RequestID: r.RequestID,
						Code:      terrors.InternalServerError,
						Data:      err,
					}
				}
				if data != nil {
					for _, entry := range data.([]*global.Entry) {
						entry.Event = s.IsEvent()
						lgs.Add(frontEntry{
							Entry: *entry,
							Time:  global.FrontUint64(entry.Time),
						})
					}
				}
			}
		}(sub)
	}

	wg.Wait()
	// fmt.Println("lgs.Len()", lgs.Len())
	result := lgs.GetAll()
	sort.Slice(result, func(i, j int) bool { return result[i].Time < result[j].Time })
	c <- &ws.ResponseS{
		RequestID: r.RequestID,
		Code:      terrors.Success,
		Data:      result,
	}

	wg = &sync.WaitGroup{}
	for _, sub := range a.Subs {
		sub.EnvID = conv.String(sub.EnvID)
		wg.Add(1)
		go func(s *Sub) {
			go s.handle(r, c)
			wg.Done()
		}(sub)
	}
	wg.Wait()
}

func (a Sub) handle(r *ws.Request, c chan<- *ws.ResponseS) {
	filter := func(*global.Entry) bool {
		return true
	}

	if a.Where != nil {
		filter = a.Where.Filter()
	}

	if a.Events {
		filter = func(e *global.Entry) bool {
			return e.Event
		}
	}

	lgs := make([]frontEntry, 0)
	ticker := time.NewTicker(100 * time.Millisecond)
	cacheSub := cache.Cache.Subscribe(r.Ctx, a.EnvID, a.Elements...)
	defer ticker.Stop()
	for {
		select {
		case <-r.Ctx.Done():
			return

		case logE := <-cacheSub:
			if filter(logE) {
				// c <- ws.Ok(r, []*global.Entry{log})
				lgs = append(lgs, frontEntry{
					Entry: *logE,
					Time:  global.FrontUint64(logE.Time),
				})
			}
			continue
		case <-ticker.C:

		}

		if len(lgs) > a.LimitRows {
			lgs = lgs[len(lgs)-a.LimitRows:]
		}

		if len(lgs) > 0 {
			c <- &ws.ResponseS{
				RequestID: r.RequestID,
				Code:      terrors.Success,
				Data:      lgs,
			}
			lgs = make([]frontEntry, 0, a.LimitRows)
		}
	}
}

func (a *Live) UnmarshalJSON(data []byte) error {
	type subAlias Sub
	type SubAction struct {
		subAlias
		Where operator.Raw
	}
	type alias Live
	var action struct {
		alias
		Subs []*SubAction
	}

	err := json.Unmarshal(data, &action)
	if err != nil {
		return err
	}

	for _, sub := range action.Subs {
		s := &Sub{
			Base:     sub.subAlias.Base,
			Elements: sub.subAlias.Elements,
			LastTime: sub.subAlias.LastTime,
		}

		err = sub.Where.Unmarshal(&s.Where)
		if err != nil {
			return err
		}

		sub.LimitRows = math.Clamp(sub.LimitRows, 1, global.JounalProxyConfig.MaxEntriesLimit)
		a.Subs = append(a.Subs, s)
	}

	return nil
}

func (s Sub) ToQuery() query.QueryI {
	return query.QueryI(&query.Vector{
		Base:      s.Base,
		Time:      global.FrontUint64(time.Now().UnixNano()),
		Direction: query.Before,
	})
}