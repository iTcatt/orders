package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(productHandler productHandler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logMiddleware())

	router.Get("/product/{id}", productHandler.GetByID)
	router.Post("/product", productHandler.Create)
	router.Patch("/product/{id}", productHandler.Update)
	router.Delete("/product/{id}", productHandler.Delete)

	return router
}

func SendJSON(w http.ResponseWriter, data any, code int) {
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
