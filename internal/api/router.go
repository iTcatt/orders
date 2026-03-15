package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/metrics"
)

func NewRouter(productHandler productHandler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logMiddleware())
	router.Use(metricsMiddleware())

	router.Route("/product", func(r chi.Router) {
		r.Get("/", productHandler.Get)
		r.Get("/{id}", productHandler.GetByID)
		r.Post("/", productHandler.Create)
		r.Patch("/{id}", productHandler.Update)
		r.Delete("/{id}", productHandler.Delete)
	})
	router.Handle("/metrics", metrics.Handler())

	return router
}
