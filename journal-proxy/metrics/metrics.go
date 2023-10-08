package metrics

import (
	"lib/utils/metrics"
	"runtime"
)

type Metrics struct {
	// system
	Sys metrics.System

	// Insert
	Interval             metrics.Value
	InsertRequests       metrics.Value
	ActiveInsertsToDB    metrics.Value
	Inserts              metrics.Value
	InsertedTotal        metrics.Value
	AvgInsertSize        metrics.Avg
	InsertRequestTime    metrics.Value `json:"InsertRequestTime(ms)"`
	AvgInsertRequestTime metrics.Avg   `json:"AvgInsertRequestTime(ms)"`

	CacheTotalTime metrics.Value `json:"CacheTotalTime(ms)"`
	AvgCacheTime   metrics.Avg   `json:"AvgCacheTime(ms)"`

	BufferSize metrics.Value
	PoolSize   metrics.Value

	// Select
	SelectRequests    metrics.Value
	SelectRequestTime metrics.Value
	AvgSelectTime     metrics.Avg
}

var Vars Metrics

func init() {
	Vars.AvgInsertRequestTime = metrics.Avg{
		Num: &Vars.InsertRequestTime,
		Div: &Vars.Inserts,
	}
	Vars.AvgSelectTime = metrics.Avg{
		Num: &Vars.SelectRequestTime,
		Div: &Vars.SelectRequests,
	}
	Vars.AvgInsertSize = metrics.Avg{
		Num: &Vars.InsertedTotal,
		Div: &Vars.Inserts,
	}
	Vars.AvgCacheTime = metrics.Avg{
		Num: &Vars.CacheTotalTime,
		Div: &Vars.Inserts,
	}
	Vars.Sys.Cpu = uint16(runtime.GOMAXPROCS(0))
}
