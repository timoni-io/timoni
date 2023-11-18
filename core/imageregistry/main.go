package imageregistry

import (
	"bytes"
	"context"
	"core/config"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/health"
	"github.com/distribution/distribution/v3/registry/handlers"
	"github.com/distribution/distribution/v3/registry/storage"
	"github.com/distribution/distribution/v3/registry/storage/driver/factory"
	"github.com/docker/go-metrics"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	// for docker registry
	_ "github.com/distribution/distribution/v3/registry/storage/driver/filesystem"
	_ "rsc.io/letsencrypt"
)

var k3sUsername string
var k3sPassword string

const imageregistryConf = `version: 0.1
log:
  fields:
    service: registry
    formatter: json
storage:
  # cache:
  #   blobdescriptor: inmemory
  filesystem:
    rootdirectory: {{DATA_PATH}}/image-registry
  delete:
    enabled: true
http:
  headers:
    X-Content-Type-Options: [nosniff]
  debug:
    prometheus:
      enabled: true
      path: "/metrics"
health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3
`

func Start(r *mux.Router) {

	config, err := configuration.Parse(bytes.NewBuffer(
		[]byte(strings.ReplaceAll(imageregistryConf, "{{DATA_PATH}}", config.DataPath())),
	))
	if err != nil {
		log.Error("error parsing configuration: ", err)
		return
	}

	// ctx := dcontext.WithVersion(context.Background(), version.Version)
	// ctx, err = configureLogging(ctx, config)
	// if err != nil {
	// 	log.Error("error configuring logger: ", err)
	// 	return
	// }

	ctx := context.Background()

	// New registry
	app := handlers.NewApp(ctx, config)
	var handler http.Handler = app

	app.RegisterHealthChecks()
	handler = alive("/", handler)
	handler = health.Handler(handler)
	handler = panicHandler(handler)

	// If registries file doesn't exist -> skip auth middleware
	if loadK3SAuthData() {
		handler = authHandler(handler)
	}

	r.PathPrefix("/v2/").Handler(handler)

	// ----------

	if config.HTTP.Debug.Prometheus.Enabled {
		path := config.HTTP.Debug.Prometheus.Path
		if path == "" {
			path = "/metrics"
		}
		log.Info("providing prometheus metrics @ ", path)
		r.PathPrefix(path).Handler(metrics.Handler())
	}
}

func alive(path string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path {
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func panicHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error(fmt.Sprintf("%v", err))
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

func logLevel(level configuration.Loglevel) log.Level {
	l, err := log.ParseLevel(string(level))
	if err != nil {
		l = log.InfoLevel
		log.Warnf("error parsing level %q: %v, using %q	", level, err, l)
	}

	return l
}

func configureLogging(ctx context.Context, config *configuration.Configuration) (context.Context, error) {
	log.SetLevel(logLevel(config.Log.Level))

	formatter := config.Log.Formatter
	if formatter == "" {
		formatter = "text" // default formatter
	}

	switch formatter {
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat:   time.RFC3339Nano,
			DisableHTMLEscape: true,
		})
	case "text":
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	default:
		// just let the library use default on empty string.
		if config.Log.Formatter != "" {
			return ctx, fmt.Errorf("unsupported logging formatter: %q", config.Log.Formatter)
		}
	}

	if config.Log.Formatter != "" {
		log.Debugf("using %q logging formatter", config.Log.Formatter)
	}

	if len(config.Log.Fields) > 0 {
		// build up the static fields, if present.
		var fields []interface{}
		for k := range config.Log.Fields {
			fields = append(fields, k)
		}

		// ctx = dcontext.WithValues(ctx, config.Log.Fields)
		// ctx = dcontext.WithLogger(ctx, dcontext.GetLogger(ctx, fields...))
	}

	// dcontext.SetDefaultLogger(dcontext.GetLogger(ctx))
	return ctx, nil
}

func CollectGarbage() {
	fp, err := os.Open("config/imageregistry.yaml")
	if err != nil {
		log.Error("error opening config/imageregistry.yaml: ", err)
		return
	}
	defer fp.Close()

	config, err := configuration.Parse(fp)
	if err != nil {
		log.Error("error parsing configuration: ", err)
		return
	}

	ctx := context.Background()
	driver, err := factory.Create(ctx, config.Storage.Type(), config.Storage.Parameters())
	if err != nil {
		log.Error(err)
		return
	}

	// k, err := libtrust.GenerateECP256PrivateKey()
	// if err != nil {
	// 	log.Error(err)
	// 	return
	// }

	// registry, err := storage.NewRegistry(ctx, driver, storage.Schema1SigningKey(k))
	registry, err := storage.NewRegistry(ctx, driver)
	if err != nil {
		log.Error(err)
		return
	}

	err = storage.MarkAndSweep(ctx, driver, registry, storage.GCOpts{
		DryRun:         false,
		RemoveUntagged: true,
	})
	if err != nil {
		log.Error(err)
		return
	}
}
