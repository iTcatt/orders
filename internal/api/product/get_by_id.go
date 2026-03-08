package product

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"iTcatt/orders/internal/api"
	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/models"
	uc "iTcatt/orders/internal/usecase/product"
)

func (h *handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	product, err := h.uc.GetProductByID(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, uc.ErrProductNotFound) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to get product by id", slog.String("err", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := convertToProduct(product)
	api.SendJSON(w, out, http.StatusOK)
}

func convertToProduct(product models.Product) dto.Product {
	return dto.Product{
		ID:          product.ID,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}
