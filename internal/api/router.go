package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	mux         *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

func NewRouter(productHandler productHandler, imageHandler imageHandler) *Router {
	prometheus.MustRegister(httpServerRequestDuration, httpServerActiveRequests)

	r := &Router{mux: http.NewServeMux()}
	r.Use(recoveryMiddleware, requestIDMiddleware, corsMiddleware, logMiddleware, metricsMiddleware)

	r.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/index.html")
	})

	r.HandleFunc("GET /product/", productHandler.Get)
	r.HandleFunc("GET /product/{id}", productHandler.GetByID)
	r.HandleFunc("POST /product/", productHandler.Create)
	r.HandleFunc("PATCH /product/{id}", productHandler.Update)
	r.HandleFunc("DELETE /product/{id}", productHandler.Delete)

	r.HandleFunc("POST /product/{id}/image", imageHandler.Upload)
	r.HandleFunc("DELETE /product/{id}/image/{imageId}", imageHandler.Delete)

	r.mux.Handle("GET /metrics", promhttp.Handler())

	return r
}

func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	handler := http.Handler(handlerFunc)
	for _, middleware := range r.middlewares {
		handler = middleware(handler)
	}
	r.mux.Handle(pattern, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
