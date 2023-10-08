package embed

import (
	"context"
	"fmt"
	"net/http"
	"time"

	// for docker registry
	"github.com/distribution/distribution/v3/configuration"
	dctx "github.com/distribution/distribution/v3/context"
	"github.com/distribution/distribution/v3/health"
	"github.com/distribution/distribution/v3/registry/handlers"
	"github.com/distribution/distribution/v3/registry/storage"
	"github.com/distribution/distribution/v3/registry/storage/driver/factory"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/filesystem"
	"github.com/distribution/distribution/v3/version"
	"github.com/docker/go-metrics"
	"github.com/docker/libtrust"
	log "github.com/sirupsen/logrus"
	_ "rsc.io/letsencrypt"
)

var (
	imgReg *ImageRegistry
)

// Start image registry endpoints.
// auth is username password pair
func Start(ctx context.Context, r *http.ServeMux, path string, auth ...string) {
	imgReg = &ImageRegistry{
		path: path,
		cfg:  CfgDefaults(path),
	}

	config, err := configuration.Parse(imgReg.cfg.Encoder())
	if err != nil {
		log.Error("error parsing configuration: ", err)
		return
	}

	ctx = dctx.WithVersion(ctx, version.Version)
	ctx, err = configureLogging(ctx, config)
	if err != nil {
		log.Error("error configuring logger: ", err)
		return
	}

	// Create registry
	app := handlers.NewApp(ctx, config)
	app.RegisterHealthChecks()

	var handler http.Handler = app
	handler = alive("/", handler)
	handler = health.Handler(handler)
	handler = panicHandler(handler)

	if len(auth) >= 2 {
		handler = authHandler(handler, auth[0], auth[1])
	}

	r.Handle("/v2/", handler)

	// ----------

	if config.HTTP.Debug.Prometheus.Enabled {
		_path := config.HTTP.Debug.Prometheus.Path
		if _path == "" {
			_path = "/metrics"
		}
		log.Info("providing prometheus metrics @ ", _path)
		r.Handle(_path, metrics.Handler())
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

		ctx = dctx.WithValues(ctx, config.Log.Fields)
		ctx = dctx.WithLogger(ctx, dctx.GetLogger(ctx, fields...))
	}

	dctx.SetDefaultLogger(dctx.GetLogger(ctx))
	return ctx, nil
}

// Garbage Collector
func CollectGarbage(ctx context.Context) {
	config, err := configuration.Parse(imgReg.cfg.Encoder())
	if err != nil {
		log.Error("error parsing configuration: ", err)
		return
	}

	driver, err := factory.Create(config.Storage.Type(), config.Storage.Parameters())
	if err != nil {
		log.Error(err)
		return
	}

	k, err := libtrust.GenerateECP256PrivateKey()
	if err != nil {
		log.Error(err)
		return
	}

	registry, err := storage.NewRegistry(ctx, driver, storage.Schema1SigningKey(k))
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
