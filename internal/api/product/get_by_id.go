package product

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"iTcatt/orders/internal/api"
	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/models"
	"iTcatt/orders/pkg/sqlp"
)

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	product, err := h.storage.GetByID(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sqlp.ErrNotFound) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to get product by id", slog.String("err", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := convertToGetProductByIDOut(product)
	api.SendJSON(w, out, http.StatusOK)
}

func convertToGetProductByIDOut(product models.Product) dto.GetProductByIDOut {
	return dto.GetProductByIDOut{
		ID:          product.ID,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}
