package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v3"
)

func logMiddleware() func(http.Handler) http.Handler {
	return httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level:         slog.LevelInfo,
		RecoverPanics: true,

		LogRequestHeaders:  []string{"Origin"},
		LogResponseHeaders: []string{},

		LogRequestBody:  func(req *http.Request) bool { return true },
		LogResponseBody: func(req *http.Request) bool { return true },
	})
}
