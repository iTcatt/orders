package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/metrics"
)

func logMiddleware() func(http.Handler) http.Handler {
	return httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level:         slog.LevelInfo,
		RecoverPanics: true,

		LogRequestHeaders:  []string{"Origin"},
		LogResponseHeaders: []string{},

		LogRequestBody:  func(req *http.Request) bool { return true },
		LogResponseBody: func(req *http.Request) bool { return true },
		Skip: func(req *http.Request, respStatus int) bool {
			return req.RequestURI == "/metrics"
		},
	})
}

func corsMiddleware() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})
}

func metricsMiddleware() func(http.Handler) http.Handler {
	return metrics.Collector(metrics.CollectorOpts{
		Host:  false,
		Proto: true,
		Skip: func(r *http.Request) bool {
			return r.RequestURI == "/metrics"
		},
	})
}
