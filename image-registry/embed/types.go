package embed

import (
	"bytes"
	"io"
	"net/http"

	"gopkg.in/yaml.v2"
)

type ImageRegistry struct {
	path string
	cfg  *ImgRegistryCfg
}

// --- Config ---

type ImgRegistryCfg struct {
	Version float64 `yaml:"version"`
	Log     Log     `yaml:"log"`
	Storage Storage `yaml:"storage"`
	HTTP    HTTP    `yaml:"http"`
	Health  Health  `yaml:"health"`
}
type Fields struct {
	Service   string `yaml:"service"`
	Formatter string `yaml:"formatter"`
}
type Log struct {
	Fields Fields `yaml:"fields"`
}
type Cache struct {
	Blobdescriptor string `yaml:"blobdescriptor"`
}
type Filesystem struct {
	Rootdirectory string `yaml:"rootdirectory"`
}
type Delete struct {
	Enabled bool `yaml:"enabled"`
}
type Storage struct {
	Cache      Cache      `yaml:"cache"`
	Filesystem Filesystem `yaml:"filesystem"`
	Delete     Delete     `yaml:"delete"`
}
type Prometheus struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}
type Debug struct {
	Prometheus Prometheus `yaml:"prometheus"`
}
type HTTP struct {
	Headers http.Header `yaml:"headers"`
	Debug   Debug       `yaml:"debug"`
}
type Storagedriver struct {
	Enabled   bool   `yaml:"enabled"`
	Interval  string `yaml:"interval"`
	Threshold int    `yaml:"threshold"`
}
type Health struct {
	Storagedriver Storagedriver `yaml:"storagedriver"`
}

func CfgDefaults(path string) *ImgRegistryCfg {
	return &ImgRegistryCfg{
		Version: 0.1,
		Log: Log{
			Fields: Fields{
				Service:   "registry",
				Formatter: "json",
			},
		},
		Storage: Storage{
			// Cache:Cache{
			// 	Blobdescriptor: "inmemory",
			// },
			Filesystem: Filesystem{
				Rootdirectory: path,
			},
			Delete: Delete{
				Enabled: true,
			},
		},
		HTTP: HTTP{
			Headers: http.Header{
				"X-Content-Type-Options": []string{"nosniff"},
			},
			Debug: Debug{
				Prometheus: Prometheus{
					Enabled: true,
					Path:    "/metrics",
				},
			},
		},
		Health: Health{
			Storagedriver: Storagedriver{
				Enabled:   true,
				Interval:  "10s",
				Threshold: 3,
			},
		},
	}
}

func (cfg ImgRegistryCfg) Encoder() io.Reader {
	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)
	defer enc.Close()

	enc.Encode(cfg)
	return buf
}
