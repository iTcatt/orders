package product

import (
	"encoding/json"
	"net/http"

	"iTcatt/orders/internal/api"
	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/usecase"
)

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	var in dto.CreateProductIn
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.uc.CreateProduct(r.Context(), convertInToDTO(in))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := dto.CreateProductOut{ID: id}
	api.SendJSON(w, out, http.StatusCreated)
}

func convertInToDTO(in dto.CreateProductIn) usecase.CreateProductIn {
	return usecase.CreateProductIn{
		Title:       in.Title,
		Description: in.Description,
		Price:       in.Price,
	}
}
