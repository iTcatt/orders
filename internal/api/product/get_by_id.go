package product

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"iTcatt/orders/internal/api"
	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
)

func (h *handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if id <= 0 {
		api.SendValidationError(w, "id must be positive")
		return
	}

	product, err := h.uc.GetProductByID(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, usecase.ErrProductNotFound) {
			api.SendNotFoundError(w, "product not found")
			return
		}

		slog.Error("failed to get product by id", slog.String("error", err.Error()))
		api.SendInternalError(w, "failed to get product by id")
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
