package product

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"iTcatt/orders/internal/api"
	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 50
)

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	in, err := extractGetInput(r)
	if err != nil {
		api.SendValidationError(w, err.Error())
		return
	}

	products, err := h.uc.GetProducts(r.Context(), in)
	if err != nil {
		slog.Error("failed to get products", slog.String("error", err.Error()))
		api.SendInternalError(w, "failed to get products")
		return
	}

	out := convertToProductSlice(products)
	api.SendJSON(w, out, http.StatusOK)
}

func extractGetInput(r *http.Request) (usecase.GetProductsIn, error) {
	page, err := parseQueryInt(r.URL.Query().Get("page"), defaultPage)
	if err != nil {
		return usecase.GetProductsIn{}, fmt.Errorf("invalid page: %w", err)
	}

	limit, err := parseQueryInt(r.URL.Query().Get("limit"), defaultLimit)
	if err != nil {
		return usecase.GetProductsIn{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return usecase.GetProductsIn{
		Page:  int32(page),
		Limit: int32(limit),
	}, nil
}

func parseQueryInt(s string, defaultVal int) (int, error) {
	if s == "" {
		return defaultVal, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("must be a positive integer")
	}
	return v, nil
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
