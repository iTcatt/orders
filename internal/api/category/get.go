package category

import (
	"log/slog"
	"net/http"

	"iTcatt/orders/internal/api/category/dto"
	"iTcatt/orders/internal/models"
	"iTcatt/orders/pkg/api"
)

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	categories, err := h.uc.GetAll(r.Context())
	if err != nil {
		slog.Error("failed to get categories", slog.Any("error", err))
		api.SendInternalError(w, "failed to get categories")
		return
	}

	api.SendJSON(w, convertToCategories(categories), http.StatusOK)
}

func convertToCategories(categories []models.Category) []dto.Category {
	out := make([]dto.Category, 0, len(categories))
	for _, c := range categories {
		out = append(out, dto.Category{ID: c.ID, Slug: c.Slug, Name: c.Name})
	}
	return out
}
