package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/metrics"
)

func NewRouter(productHandler productHandler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		corsMiddleware(),
		logMiddleware(),
		metricsMiddleware(),
	)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/index.html")
	})

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
