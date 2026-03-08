package api

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter(productHandler productHandler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logMiddleware())

	router.Route("/product", func(r chi.Router) {
		r.Get("/", productHandler.Get)
		r.Get("/{id}", productHandler.GetByID)
		r.Post("/", productHandler.Create)
		r.Patch("/{id}", productHandler.Update)
		r.Delete("/{id}", productHandler.Delete)
	})

	return router
}
