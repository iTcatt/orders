package product

import (
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
	in := extractGetInput(r)

	products, err := h.uc.GetProducts(r.Context(), in)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := convertToProductSlice(products)
	api.SendJSON(w, out, http.StatusOK)
}

func extractGetInput(r *http.Request) usecase.GetProductsIn {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = defaultPage
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return usecase.GetProductsIn{
		Page:  int32(page),
		Limit: int32(limit),
	}
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
