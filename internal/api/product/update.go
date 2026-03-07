package product

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/internal/usecase/product"
)

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, in, err := extractInput(r)
	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	if err := h.uc.UpdateProduct(r.Context(), id, in); err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func extractInput(r *http.Request) (int32, usecase.UpdateProductIn, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, usecase.UpdateProductIn{}, fmt.Errorf("invalid id: %w", err)
	}

	var in dto.UpdateProductIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return 0, usecase.UpdateProductIn{}, fmt.Errorf("invalid input: %w", err)
	}

	return int32(id), usecase.UpdateProductIn{
		Title:       in.Title,
		Description: in.Description,
		Price:       in.Price,
	}, nil
}
