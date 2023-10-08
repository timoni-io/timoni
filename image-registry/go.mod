module image-registry

go 1.20

require (
	github.com/distribution/distribution/v3 v3.0.0-20230627210631-71a6c56fbf6c
	github.com/docker/go-metrics v0.0.1
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/yaml.v2 v2.4.0
	lib v0.0.0-00010101000000-000000000000
	rsc.io/letsencrypt v0.0.3
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gomodule/redigo v1.8.2 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/klauspost/compress v1.16.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	golang.org/x/sys v0.9.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)

replace lib => ../lib
