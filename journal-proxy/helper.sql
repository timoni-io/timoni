DROP DATABASE IF EXISTS logs ON CLUSTER 'timoni';
CREATE DATABASE logs ON CLUSTER 'timoni';


DROP TABLE IF EXISTS logs.test_table ON CLUSTER 'timoni';

CREATE TABLE IF NOT EXISTS logs.test_table ON CLUSTER 'timoni' (
    id UUID MATERIALIZED generateUUIDv4(),
    message String,
    time DateTime64(9) DEFAULT now()
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{installation}/{cluster}/{database}/{table}/{shard}', '{replica}')
ORDER BY(time)
SETTINGS index_granularity = 32768;

-- node 1
INSERT INTO logs.test_table (message) VALUES ('a'), ('b');
-- node 2
INSERT INTO logs.test_table (message) VALUES ('c'), ('d');

SELECT * FROM logs.test_table;


DROP TABLE IF EXISTS logs.dist_table ON CLUSTER 'timoni';

CREATE TABLE IF NOT EXISTS logs.dist_table ON CLUSTER 'timoni'
ENGINE = Distributed('timoni', logs, test_table, rand());

-- node 1
INSERT INTO logs.dist_table (message) VALUES ('a'), ('b');
-- node 2
INSERT INTO logs.dist_table (message) VALUES ('c'), ('d');

SELECT * FROM logs.test_table;
SELECT * FROM logs.dist_table;
