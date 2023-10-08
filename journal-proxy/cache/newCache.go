package cache

import (
	"context"
	"journal-proxy/global"
	"lib/storage"
	"lib/storage/helpers"
	"lib/storage/memory"
	"lib/utils"
	"lib/utils/iter"
	"lib/utils/slice"
	"lib/utils/types"
)

// New cache using lib/storage
// TODO: Finish or delete it

const (
	entryKey = "entries"
)

type cache struct {
	data storage.Connection
}

func NewCache() *cache {
	return &cache{
		data: utils.Must(memory.New()),
	}
}

func (c *cache) Init(entries []*global.Entry, tags map[string]CacheTags) {
	c.Update(entries)
}

func (c *cache) Update(entries []*global.Entry) {
	logEntries := map[string][]*global.Entry{}
	for _, entry := range entries {
		logEntries[entry.EnvID] = append(logEntries[entry.EnvID], entry)
	}

	for envID, entries := range logEntries {
		c.data.Bucket(envID).Tx(func(tx storage.Transactioner) error {
			// update entries
			logs, _ := helpers.Get[[]*global.Entry](tx, entryKey)
			tx.Set(entryKey, append(logs, entries...))
			return nil
		})
	}
}

// TODO: implement Tags cache
func (*cache) Tags(envID, key string) []string {
	return nil
}

func (c *cache) Entries(envID string, elements ...string) []*global.Entry {
	entries, err := helpers.Get[[]*global.Entry](c.data.Bucket(envID), entryKey)
	if err != nil {
		return nil
	}

	if len(elements) == 0 {
		return entries
	}

	return iter.FilterSlice(entries, func(entry *global.Entry) bool {
		return slice.Contains(elements, entry.Element)
	})
}

func (c *cache) Subscribe(ctx context.Context, envID string, elements ...string) <-chan *global.Entry {
	out := make(chan *global.Entry, 100)
	go func() {
		defer close(out)
		watch := helpers.Watch[[]*global.Entry](ctx, c.data.Bucket(envID), entryKey)

		for i := range watch {
			if i.Event != types.PutEvent {
				continue
			}
			for _, entry := range i.Value {
				if len(elements) > 0 && !slice.Contains(elements, entry.Element) {
					continue
				}
				out <- entry
			}
		}
	}()
	return out
}
