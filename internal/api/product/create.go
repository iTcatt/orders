package product

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"time"

	"iTcatt/orders/internal/api"
	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/models"
)

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var in dto.CreateProductIn
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	product := convertInToModel(in)
	err = h.storage.Create(r.Context(), product)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := dto.CreateProductOut{ID: product.ID}
	api.SendJSON(w, out, http.StatusCreated)
}

func convertInToModel(in dto.CreateProductIn) models.Product {
	return models.Product{
		ID:          rand.Int32(),
		Title:       in.Title,
		Description: in.Description,
		Price:       in.Price,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
