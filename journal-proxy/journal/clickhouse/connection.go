package clickhouse

import (
	"context"
	"fmt"
	"journal-proxy/journal"
	"lib/utils/set"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ConnectionS struct {
	CTX      context.Context
	DB       clickhouse.Conn
	uri      string
	tableSet *set.Safe[string]
}

func NewConnection(addr string) journal.ConnectionI {

	if len(strings.Split(addr, ":")[0]) == 0 {
		return nil
	}

	c := &ConnectionS{
		uri:      addr,
		CTX:      context.Background(),
		tableSet: set.NewSafe[string](nil),
	}

	c.Connect()

	if !c.Connected() {
		return c
	}

	c.tableSet.Add(tablesToEnvNames(c.getExistsingTables()...)...)

	return c
}

func (c *ConnectionS) Connected() bool {
	return c.DB != nil && c.DB.Ping(c.CTX) == nil
}

func (c *ConnectionS) Connect() error {
	if c.Connected() {
		return nil
	}

	db, _ := clickhouse.Open(&clickhouse.Options{
		Interface: clickhouse.NativeInterface,
		Addr:      []string{c.uri},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "timoni",
			Password: "timoni",
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
			"log_queries":        0,
			"log_query_threads":  0,
			"connect_timeout":    3,
		},
		DialTimeout: 10 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		// Debug: true,
	})

	err := db.Ping(c.CTX)
	if err != nil || db == nil {
		return fmt.Errorf("connection failed %s: %w", c.uri, err)
	}

	c.DB = db
	return nil
}

func (c *ConnectionS) GetExistingEnvs() []string {
	return c.tableSet.List()
}
