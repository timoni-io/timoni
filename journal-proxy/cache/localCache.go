package cache

import (
	"context"
	"fmt"
	"journal-proxy/global"
	"lib/tlog"
	"lib/utils/channel"
	"lib/utils/maps"
	"lib/utils/set"
	"lib/utils/slice"
	"sort"
	"strings"
	"sync"
	"time"
)

type localCache struct {
	// envID -> tagCache
	EnvTags maps.Maper[string, *tagCache]
	// envID -> element name -> entryCache
	// EnvEntries maps.Maper[string, maps.Maper[string, *entryCache]]
	EnvEntries maps.Maper[string, *envCache]

	valuesLimit uint16
	entryLimit  uint16
}

type envCache struct {
	maps.Maper[string, *entryCache]
	Live *channel.Hub[*global.Entry]
}

type tagCache struct {
	Keys *set.Safe[string]
	// tag name -> RigidSet of values
	ValuesStr maps.Maper[string, *set.Safe[string]]
	// tag name -> RigidSet of values
	ValuesNum maps.Maper[string, *set.Safe[float64]]
}

type entryCache struct {
	// element name -> RigidSlice of entries
	Entries *slice.Rigid[*global.Entry, uint16]
	Broker  *channel.Hub[*global.Entry]
}

func newEnvCache() *envCache {
	return &envCache{
		maps.New[string, *entryCache](nil).Safe(),
		channel.NewHub[*global.Entry](context.Background(), 30),
	}
}
func newTagCache(valuesLimit uint16) *tagCache {
	return &tagCache{
		Keys:      set.NewSafe[string](nil),
		ValuesStr: maps.New(map[string]*set.Safe[string]{}).Safe(),
		ValuesNum: maps.New(map[string]*set.Safe[float64]{}).Safe(),
	}
}

func newEntryCache() *entryCache {
	return &entryCache{
		Entries: slice.NewRigid[*global.Entry](global.JounalProxyConfig.CacheEntryLimit),
		Broker:  channel.NewHub[*global.Entry](context.TODO(), 1<<15),
	}
}

func NewLocal(valuesLimit, entryLimit uint16) CacheI {
	return &localCache{
		EnvTags:    maps.New[string, *tagCache](nil).Safe(),
		EnvEntries: maps.New[string, *envCache](nil).Safe(),

		valuesLimit: valuesLimit,
		entryLimit:  entryLimit,
	}
}

func (c *localCache) Init(entries []*global.Entry, tags map[string]CacheTags) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		c.updateEntryCache(entries)
		wg.Done()
	}()

	for envID, tags := range tags {
		wg.Add(1)
		go func(envID string, tags CacheTags) {
			defer wg.Done()

			tagsCache, ok := c.EnvTags.GetFull(envID)
			if !ok {
				tagsCache = newTagCache(c.valuesLimit)
				c.EnvTags.Set(envID, tagsCache)
			}
			tagsCache.Keys.Add(tags.Keys...)

			for tagName, tagValues := range tags.Strings {
				tagsCache.ValuesStr.Commit(func(data map[string]*set.Safe[string]) {
					x, ok := data[tagName]
					if !ok {
						x = set.NewRigid[string](global.JounalProxyConfig.CacheValuesLimit).Safe()
						data[tagName] = x
					}
					x.Add(tagValues...)
				})
			}

			for tagName, tagValues := range tags.Numbers {
				tagsCache.ValuesNum.Commit(func(data map[string]*set.Safe[float64]) {
					x, ok := data[tagName]
					if !ok {
						x = set.NewRigid[float64](global.JounalProxyConfig.CacheValuesLimit).Safe()
						data[tagName] = x
					}
					x.Add(tagValues...)
				})
			}
		}(envID, tags)
	}

	wg.Wait()
}

func (c *localCache) Exists(envID string) bool {
	_, ok := c.EnvEntries.GetFull(envID)
	return ok
}

func (c *localCache) Update(entries []*global.Entry) {

	// fmt.Printf("Update %+v\n", entries[0])

	wg := sync.WaitGroup{}
	wg.Add(2)

	// cache entry
	go func() {
		defer wg.Done()
		c.updateTagCache(entries)
	}()

	go func() {
		defer wg.Done()
		c.updateEntryCache(entries)
	}()

	wg.Wait()
}

func (c *localCache) Tags(envid, key string) []string {
	tc := c.EnvTags.Get(envid)
	if key == "" {
		return tc.Keys.List()
	}

	vStr, isStr := tc.ValuesStr.GetFull(key)
	vNum, isNum := tc.ValuesNum.GetFull(key)
	switch {
	case isStr:
		return vStr.List()
	case isNum:
		var x []string
		for _, v := range vNum.List() {
			x = append(x, fmt.Sprintf("%f", v))
		}
		return x
	default:
		return nil
	}
}

func (c *localCache) Entries(envID string, elements ...string) []*global.Entry {
	envCch := c.EnvEntries.Get(envID)
	if envCch == nil {
		return nil
	}

	if len(elements) == 0 {
		elements = envCch.Keys()
	}

	entries := make([]*global.Entry, 0, global.JounalProxyConfig.CacheEntryLimit*uint16(len(elements)))
	for _, element := range elements {
		elementCache := envCch.Get(element)
		if elementCache == nil {
			continue
		}
		entries = append(entries, elementCache.Entries.GetAll()...)
	}

	// sort from oldest to newest
	sort.Slice(entries, func(i, j int) bool { return entries[i].Time < entries[j].Time })

	if len(entries) > global.JounalProxyConfig.MaxEntriesLimit {
		return entries[len(entries)-global.JounalProxyConfig.MaxEntriesLimit:]
	}
	return entries
}

func (c *localCache) updateEntryCache(entries []*global.Entry) {
	if c == nil {
		tlog.Error("c *localCache is nil")
		return
	}

	envs := map[string][]*global.Entry{}
	for _, entry := range entries {
		envs[entry.EnvID] = append(envs[entry.EnvID], entry)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(envs))
	for envID, entries := range envs {
		go func(envID string, entries []*global.Entry) {
			defer wg.Done()

			var ok bool
			var envCch *envCache
			c.EnvEntries.Commit(func(data map[string]*envCache) {
				envCch, ok = data[envID]
				if !ok {
					envCch = newEnvCache()
					data[envID] = envCch
				}
			})

			envCch.Commit(func(elementCache map[string]*entryCache) {
				for _, entry := range entries {
					entryCch, ok := elementCache[entry.Element]
					if !ok {
						entryCch = newEntryCache()
						elementCache[entry.Element] = entryCch
					}
					entryCch.Entries.Add(entry)
					entryCch.Broker.Broadcast(entry)
					envCch.Live.Broadcast(entry)
				}
			})
		}(envID, entries)
	}

	wg.Wait()
}

func (c *localCache) updateTagCache(entries []*global.Entry) {
	envs := map[string][]*global.Entry{}
	for _, entry := range entries {
		envs[entry.EnvID] = append(envs[entry.EnvID], entry)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(envs))
	for envID, entries := range envs {
		go func(envID string, entries []*global.Entry) {
			defer wg.Done()

			tc, ok := c.EnvTags.GetFull(envID)
			if !ok {
				tc = newTagCache(c.valuesLimit)
				c.EnvTags.Set(envID, tc)
			}

			for _, entry := range entries {

				for tagName, tagValue := range entry.TagsString {
					tagName = strings.TrimSpace(tagName)
					tagValue = strings.TrimSpace(tagValue)
					tc.Keys.Add(tagName)

					tc.ValuesStr.Commit(func(data map[string]*set.Safe[string]) {
						x, ok := data[tagName]
						if !ok {
							x = set.NewRigid[string](global.JounalProxyConfig.CacheValuesLimit).Safe()
							data[tagName] = x
						}
						x.Add(tagValue)
					})
				}

				for tagName, tagValue := range entry.TagsNumber {
					tagName = strings.TrimSpace(tagName)
					tc.Keys.Add(tagName)

					tc.ValuesNum.Commit(func(data map[string]*set.Safe[float64]) {
						x, ok := data[tagName]
						if !ok {
							x = set.NewRigid[float64](global.JounalProxyConfig.CacheValuesLimit).Safe()
							data[tagName] = x
						}
						x.Add(tagValue)
					})
				}

			} // for entry in entries
		}(envID, entries)
	}

	wg.Wait()
}

func (l *localCache) Subscribe(ctx context.Context, envID string, elements ...string) <-chan *global.Entry {
	envCach := l.EnvEntries.Get(envID)
	if envCach == nil {
		l.EnvEntries.Commit(func(data map[string]*envCache) {
			_, ok := data[envID]
			if !ok {
				envCach = newEnvCache()
				data[envID] = envCach
			}
		})
	}
	if len(elements) == 0 {
		return envCach.Live.Register(ctx)
	}

	if len(elements) == 1 {
		// front-end might ask for a element that is still building or starting up
		// so we need to wait for it to be ready
		for i := 0; i < 25; i++ {
			elementCache := envCach.Get(elements[0])
			if elementCache != nil {
				c := elementCache.Broker.Register(ctx)
				return c
			}
			time.Sleep(time.Second)
		}
		return nil
	}

	channels := make([]<-chan *global.Entry, len(elements))

	for i, element := range elements {
		elementCache := envCach.Get(element)
		if elementCache == nil {
			continue
		}

		c := elementCache.Broker.Register(ctx)
		channels[i] = c
	}

	return channel.Join(channels...)
}
