module journal-proxy

go 1.20

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.2.0
	github.com/buger/jsonparser v1.1.1
	github.com/gorilla/websocket v1.5.0
	golang.org/x/exp v0.0.0-20230626212559-97b1e661b5df
	lib v0.0.0-00010101000000-000000000000
)

require (
	github.com/barkimedes/go-deepcopy v0.0.0-20220514131651-17c30cfc62df // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/paulmach/orb v0.7.1 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/otel v1.10.0 // indirect
	go.opentelemetry.io/otel/trace v1.10.0 // indirect
)

replace lib => ../lib
