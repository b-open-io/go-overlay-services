package main

import (
	"net/http"

	"github.com/4chain-ag/go-overlay-services/pkg/config"
	"github.com/4chain-ag/go-overlay-services/pkg/server"
	"github.com/gookit/slog"
)

func main() {
	loader := config.NewLoader("OVERLAY")
	cfg, err := loader.Load()
	if err != nil {
		slog.Fatalf("failed to load config: %v", err)
	}

	opts := []server.HTTPOption{
		server.WithConfig(&server.Config{
			Addr: cfg.Address,
			Port: cfg.Port,
		}),
		server.WithMiddleware(loggingMiddleware),
	}

	httpAPI := server.New(opts...)

	slog.Infof("Starting server on %s:%d...", cfg.Address, cfg.Port)
	if err := httpAPI.ListenAndServe(); err != nil {
		slog.Fatalf("HTTP server failed: %v", err)
	}
}

// loggingMiddleware is a custom definition of the logging middleware format accepted by the HTTP API.
func loggingMiddleware(next http.Handler) http.Handler {
	slog.SetLogLevel(slog.DebugLevel)
	slog.SetFormatter(slog.NewJSONFormatter(func(f *slog.JSONFormatter) {
		f.PrettyPrint = true
	}))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.WithFields(slog.M{
			"category":    "service",
			"method":      r.Method,
			"remote-addr": r.RemoteAddr,
			"request-uri": r.RequestURI,
		})
		logger.Info("log-line")
		next.ServeHTTP(w, r)
	})
}
