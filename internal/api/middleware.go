package api

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	pkgapi "iTcatt/orders/pkg/api"
)

type contextKey string

const (
	maxBodyLogSize            = 4 * 1024 // 4KB
	requestIDKey   contextKey = "request_id"
)

var (
	httpServerRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_server_request_duration_seconds",
			Help:    "Duration of HTTP server requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"http_request_method", "http_route", "http_response_status_code"},
	)
	httpServerActiveRequests = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_server_active_requests",
		Help: "Number of active HTTP server requests.",
	})
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func requestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 64<<10) // 64KB - stack size
				buf = buf[:runtime.Stack(buf, false)]
				slog.Error("panic recovered",
					slog.Any("error", err),
					slog.String("stack", string(buf)),
				)
				pkgapi.SendInternalError(w, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" || r.URL.Path == "/" {
			next.ServeHTTP(w, r)
			return
		}

		var body string
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
			body = readBody(r)
		}

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rw, r)

		attrs := []any{
			slog.String("request_id", requestIDFromContext(r.Context())),
			slog.String("method", r.Method),
			slog.String("url", r.URL.RequestURI()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.Int("status", rw.statusCode),
			slog.String("proto", r.Proto),
			slog.Duration("duration", time.Since(start)),
		}
		if body != "" {
			attrs = append(attrs, slog.String("body", body))
		}

		switch {
		case rw.statusCode >= 500:
			slog.Error("request", attrs...)
		case rw.statusCode >= 400:
			slog.Warn("request", attrs...)
		default:
			slog.Info("request", attrs...)
		}
	})
}

func readBody(r *http.Request) string {
	if r.Body == nil {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyLogSize))
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))

	if err != nil || len(body) == 0 {
		return ""
	}
	return string(body)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" || r.URL.Path == "/" {
			next.ServeHTTP(w, r)
			return
		}
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		httpServerActiveRequests.Inc()
		defer httpServerActiveRequests.Dec()
		start := time.Now()
		next.ServeHTTP(rw, r)
		httpServerRequestDuration.WithLabelValues(
			r.Method,
			r.Pattern,
			strconv.Itoa(rw.statusCode),
		).Observe(time.Since(start).Seconds())
	})
}
