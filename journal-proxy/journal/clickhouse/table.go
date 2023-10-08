package clickhouse

import (
	"fmt"
	"strings"

	"lib/tlog"
)

func tablesToEnvNames(tables ...string) []string {
	var envNames []string
	for _, table := range tables {
		if !strings.HasPrefix(table, "logs_") {
			continue
		}
		envNames = append(envNames, strings.TrimPrefix(table, "logs_"))
	}
	return envNames
}

func (c *ConnectionS) CreateTables(envID string) {

	if envID == "" {
		tlog.Error("envID is empty")
		return
	}

	if c.tableSet.Exists(envID) {
		return
	}

	err := c.DB.Exec(c.CTX, "CREATE DATABASE IF NOT EXISTS logs ON CLUSTER 'timoni';")
	if err != nil {
		tlog.Error(err)
		return
	}

	for _, v := range []string{"events_%s", "logs_%s"} {
		tableName := fmt.Sprintf(v, envID)

		tableLocal := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS logs.local_%s ON CLUSTER 'timoni' (
				time UInt64,
				level String,
				message String,

				env_id String,
				element String,
				pod String,
				version String,
				git_repo String,
				user_email String,

				tags_string Map(String, String),
				tags_number Map(String, Float64)
			)
			ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{installation}/{cluster}/{database}/{table}/{shard}', '{replica}')
			PARTITION BY toYYYYMMDD(fromUnixTimestamp64Nano(time))
			ORDER BY(time, element)
			PRIMARY KEY time;
		`, tableName)

		tableDist := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS logs.%[1]s ON CLUSTER 'timoni'
			ENGINE = Distributed('timoni', logs, local_%[1]s, rand());
		`, tableName)

		err = c.DB.Exec(c.CTX, tableLocal)
		if err != nil {
			tlog.Error(err, tlog.Vars{
				"envID":     envID,
				"tableName": tableName,
			})
			return
		}

		err = c.DB.Exec(c.CTX, tableDist)
		if err != nil {
			tlog.Error(err, tlog.Vars{
				"envID":     envID,
				"tableName": tableName,
			})
			return
		}

	}
	c.tableSet.Add(envID)
}

func (c *ConnectionS) getExistsingTables() []string {
	var tables []struct {
		Names []string
	}
	err := c.DB.Select(c.CTX, &tables, "SELECT groupArray(name) AS Names FROM system.tables WHERE database = 'logs'")
	if err != nil {
		// database does not exist
		tlog.Error(err)
		return nil
	}
	return tables[0].Names
}

func (c *ConnectionS) DropTables(envID string) {
	c.tableSet.Delete(envID)

	for _, v := range []string{"events_%s", "logs_%s"} {
		sql := fmt.Sprintf("DROP TABLE IF EXISTS logs.local_%s ON CLUSTER 'timoni'", v)
		err := c.DB.Exec(c.CTX, sql)
		tlog.Error(err)

		sql = fmt.Sprintf("DROP TABLE IF EXISTS logs.%s ON CLUSTER 'timoni'", v)
		err = c.DB.Exec(c.CTX, sql)
		tlog.Error(err)
	}
}
