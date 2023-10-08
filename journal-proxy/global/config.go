package global

import (
	"encoding/base64"
	"encoding/json"
	"lib/tlog"
	"os"
)

var (
	GitTag = "???"

	JounalProxyConfig = func() JounalProxyS {
		conf := new(JounalProxyS)
		buf, _ := base64.StdEncoding.DecodeString(os.Getenv("CONF"))
		err := json.Unmarshal(buf, conf)
		tlog.Fatal(err)

		if conf.MaxEntriesLimit == 0 {
			conf = &JounalProxyS{
				MaxEntriesLimit:  1000,
				CacheValuesLimit: 1000,
				CacheEntryLimit:  1000,

				Name:                conf.Name,
				DatabaseConnections: conf.DatabaseConnections,
				DatabaseAddress:     conf.DatabaseAddress,
			}
		}

		return *conf
	}()
)

type JounalProxyS struct {
	Name  string // JP name/description
	Count int32

	DatabaseConnections int `max:"10"` // no more than number of CPU cores
	DatabaseAddress     string

	CacheValuesLimit uint16
	CacheEntryLimit  uint16
	MaxEntriesLimit  int
}
