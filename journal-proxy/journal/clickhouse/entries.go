package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"journal-proxy/global"
	"lib/tlog"
	"lib/utils/conv"
	"strings"
	"sync"
	"time"
)

func (c *ConnectionS) insert(tableName string, entries ...*global.Entry) error {

	c.CreateTables(entries[0].EnvID)
	batch, err := c.DB.PrepareBatch(context.Background(), fmt.Sprintf("insert into logs.%s", tableName))
	if err != nil {
		tlog.Error(err)
		return err
	}
	for _, e := range entries {
		if err := batch.AppendStruct(e); err != nil {
			tlog.Error(err)
			tlog.PrintJSON(e)
		}
	}

	err = batch.Send()
	if err != nil {
		if strings.Contains(err.Error(), "code: 60") {
			c.tableSet.Delete(tablesToEnvNames(tableName)...)
		}
		tlog.Error(err)
		return err
	}

	return nil
}


// Insert single entry
func (c *ConnectionS) InsertOne(entry *global.Entry) error {
	if entry == nil {
		return errors.New("nil entry")
	}
	return c.insert(prepareEntry(entry), entry)
}

// Insert multiple entries
func (c *ConnectionS) InsertMulti(entries []*global.Entry) {
	groups := map[string][]*global.Entry{}

	// Create db table groups
	for _, entry := range entries {
		if entry == nil {
			continue
		}

		tableName := prepareEntry(entry)
		groups[tableName] = append(groups[tableName], entry)
	}

	// Insert groups
	wg := sync.WaitGroup{}
	for tableName, entries := range groups {
		wg.Add(1)

		go func(tableName string, entries []*global.Entry) {
			defer wg.Done()
			c.insert(tableName, entries...)
		}(tableName, entries)
	}
	wg.Wait()
}

func prepareEntry(entry *global.Entry) (tableName string) {
	entry.EnvID = conv.String(entry.EnvID)
	entry.Message = strings.TrimSpace(entry.Message)
	if entry.Time == 0 {
		entry.Time = uint64(time.Now().UnixNano())
	}

	tablePrefix := "logs"
	if entry.Event {
		tablePrefix = "events"
		entry.TagsString["event"] = "true"
	}
	return fmt.Sprintf("%s_%s", tablePrefix, entry.EnvID)
}
