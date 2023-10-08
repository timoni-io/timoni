package cache

import (
	"context"
	"journal-proxy/global"
	"journal-proxy/journal"
	"journal-proxy/journal/query"
	"lib/tlog"
	"lib/utils/set"
	"time"
)

type CacheTags struct {
	Keys    []string
	Strings map[string][]string  // key: tag name
	Numbers map[string][]float64 // key: tag name
}

type CacheI interface {
	// tags key: envID
	Init(entries []*global.Entry, tags map[string]CacheTags)
	Update(entries []*global.Entry)

	Tags(envID, key string) []string
	Entries(envID string, elements ...string) []*global.Entry

	Subscribe(ctx context.Context, envID string, elements ...string) <-chan *global.Entry
}

// var Cache = NewCache()
var Cache = NewLocal(global.JounalProxyConfig.CacheValuesLimit, global.JounalProxyConfig.CacheEntryLimit)

func Init(conn journal.ConnectionI) journal.ConnectionI {
	if Cache == nil {
		tlog.Fatal("Cache is nil")
	}

	tlog.Info("Initializing cache")

	base := query.Base{
		EnvID:     "",
		LimitRows: 50,
		FullLog:   true,
	}

	entriesQ := &query.Vector{
		Base:      base,
		Time:      global.FrontUint64(time.Now().UnixNano()),
		Direction: query.Before,
	}

	// tagsQ := &query.TagsInit{
	// 	Base: base,
	// }

	var entries []*global.Entry
	envCacheMap := map[string]CacheTags{}

	for _, envID := range conn.GetExistingEnvs() {
		entriesQ.Base.EnvID = envID

		ctx := context.Background()

		// Get first n logs
		data, err := conn.ExecQuery(ctx, entriesQ)
		envEntries, ok := data.([]*global.Entry)
		if !ok || err != nil {
			continue
		}
		entries = append(entries, envEntries...)

		// Get all tags
		// data, err = conn.ExecQuery(ctx, tagsQ)
		envTags, ok := data.(*journal.TagInitU)
		if !ok || err != nil {
			continue
		}
		keys := set.Set[string]{}
		for key := range envTags.Strings {
			keys.Add(key)
		}
		for key := range envTags.Numbers {
			keys.Add(key)
		}
		envCacheMap[envID] = CacheTags{
			Keys:    keys.List(),
			Strings: envTags.Strings,
			Numbers: envTags.Numbers,
		}
	}

	Cache.Init(entries, envCacheMap)
	tlog.Info("Cache initialized")
	return conn
}
