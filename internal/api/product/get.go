package product

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/pkg/api"
)

const (
	defaultPage  uint32 = 1
	defaultLimit uint32 = 10
	maxLimit     uint32 = 50
)

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	in, err := extractGetInput(r)
	if err != nil {
		api.SendValidationError(w, err.Error())
		return
	}

	products, err := h.uc.GetProducts(r.Context(), in)
	if err != nil {
		slog.Error("failed to get products", slog.Any("error", err))
		api.SendInternalError(w, "failed to get products")
		return
	}

	out := convertToProductSlice(products)
	api.SendJSON(w, out, http.StatusOK)
}

func extractGetInput(r *http.Request) (usecase.GetProductsIn, error) {
	page, err := parseQueryUint32(r.URL.Query().Get("page"), defaultPage)
	if err != nil {
		return usecase.GetProductsIn{}, fmt.Errorf("invalid page: %w", err)
	}

	limit, err := parseQueryUint32(r.URL.Query().Get("limit"), defaultLimit)
	if err != nil {
		return usecase.GetProductsIn{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return usecase.GetProductsIn{
		Page:  page,
		Limit: limit,
	}, nil
}

func parseQueryUint32(s string, defaultVal uint32) (uint32, error) {
	if s == "" {
		return defaultVal, nil
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil || v == 0 {
		return 0, fmt.Errorf("must be a positive integer")
	}
	return uint32(v), nil
}

func convertToProductSlice(products []models.Product) []dto.Product {
	out := make([]dto.Product, 0, len(products))
	for _, product := range products {
		out = append(out, dto.Product{
			ID:          product.ID,
			Title:       product.Title,
			Description: product.Description,
			Price:       product.Price,
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		})
	}
	return out
}
