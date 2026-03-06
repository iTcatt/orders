package api

import (
	"encoding/json"
	"net/http"
)

func NewRouter(productHandler productHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /product/{id}", productHandler.GetByID)
	mux.HandleFunc("POST /product", productHandler.Create)
	mux.HandleFunc("PATCH /product/{id}", productHandler.Update)
	mux.HandleFunc("DELETE /product/{id}", productHandler.Delete)

	return mux
}

func SendJSON(w http.ResponseWriter, data any, code int) {
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
