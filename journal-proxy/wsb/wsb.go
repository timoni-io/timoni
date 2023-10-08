package wsb

import (
	"encoding/gob"
	"journal-proxy/cache"
	"journal-proxy/global"
	"journal-proxy/journal"
	"journal-proxy/metrics"
	"journal-proxy/parser"
	"lib/tlog"
	"lib/utils"
	"lib/utils/net"
	"lib/utils/slice"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type MessageS struct {
	global.Message
}

const (
	buffersize = 150 << 10
)

var (
	interval = 100
	buffer   = slice.NewSlice[*global.Entry](buffersize)
	ticker   = time.NewTicker(time.Millisecond * time.Duration(interval))

	ConnPool = utils.NewPool[journal.ConnectionI](global.JounalProxyConfig.DatabaseConnections)
	logPool  = make(chan []*global.Entry, 1<<7)

	upgrader = websocket.Upgrader{
		CheckOrigin:       func(r *http.Request) bool { return true },
		EnableCompression: true,
	}
)

func Loop() {
	for {
		select {
		case entries := <-logPool:
			go func(entries []*global.Entry) {
				conn := ConnPool.GetNoWait()
				insert(entries, conn)
				if conn != nil {
					ConnPool.Add(conn)
				}
			}(entries)

		case <-ticker.C:
			if buffer.Len() > 0 {
				metrics.Vars.BufferSize.Store(int64(buffer.Len()))
				logPool <- buffer.Take()
				metrics.Vars.PoolSize.Add(1)
				metrics.Vars.Interval.Store(int64(interval))

				if metrics.Vars.PoolSize.Load() > 10 {
					ticker.Stop()
					interval += 100
					ticker = time.NewTicker(time.Millisecond * time.Duration(interval))
					metrics.Vars.Interval.Store(int64(interval))

				} else if interval > 100 && metrics.Vars.PoolSize.Load() < 5 {
					ticker.Stop()
					interval -= 100
					ticker = time.NewTicker(time.Millisecond * time.Duration(interval))
					metrics.Vars.Interval.Store(int64(interval))
				}
			}
		}
	}
}

func insert(entries []*global.Entry, conn journal.ConnectionI) {
	metrics.Vars.PoolSize.Add(-1)
	metrics.Vars.ActiveInsertsToDB.Add(1)

	size := int64(len(entries))

	if size > 0 {
		wg := sync.WaitGroup{}

		// Update cache
		wg.Add(1)
		go func() {
			defer wg.Done()
			ts := time.Now()
			cache.Cache.Update(entries)
			metrics.Vars.CacheTotalTime.Add(int64(time.Since(ts).Milliseconds()))
		}()

		// Insert to DB
		if conn != nil && conn.Connected() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ts := time.Now()
				conn.InsertMulti(entries)
				metrics.Vars.InsertRequestTime.Add(int64(time.Since(ts).Milliseconds()))
			}()
		}

		wg.Wait()
		metrics.Vars.Inserts.Add(1)
		metrics.Vars.InsertedTotal.Add(size)
		metrics.Vars.ActiveInsertsToDB.Add(-1)
	}
}

func (msg *MessageS) handle(conn *websocket.Conn) {

	metrics.Vars.InsertRequests.Add(1)

	entries := parser.Parse(msg.Message)
	if len(entries) == 0 {
		tlog.Error("No entries found in request, conn: " + conn.RemoteAddr().String())
		return
	}

	buffer.Commit(func(data *[]*global.Entry, capacity int) {
		*data = append(*data, entries...)

		if len(*data) > buffersize {
			logPool <- *data
			metrics.Vars.PoolSize.Add(1)

			*data = make([]*global.Entry, 0, capacity)
			ticker.Reset(time.Millisecond * time.Duration(interval))
		}
	})
	// messagePool.Put(msg)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Print new connection info
	tlog.Info("New connection " + net.RequestIP(r))

	// Upgrade raw HTTP connection to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		tlog.Error("Error during connection upgrade:", err)
		return
	}

	// Start ws handler
	go func() {
		for {
			msg := &MessageS{}
			_, rd, err := conn.NextReader()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					tlog.Error("Error during message reading:", err)
					conn.Close()
				}
				break
			}

			err = gob.NewDecoder(rd).Decode(msg)
			if err != nil {
				tlog.Error("Error during message decoding:", err)
				break
			}

			// Handle
			go msg.handle(conn)
		}
	}()
}
