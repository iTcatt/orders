package product

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/pkg/api"
)

func (h *handler) GetByID(w http.ResponseWriter, r *http.Request) {
	parsed, err := uuid.Parse(r.PathValue("id"))
	if err != nil || parsed.Version() != 7 {
		api.SendValidationError(w, "id must be a valid UUID v7")
		return
	}

	product, err := h.uc.GetProductByID(r.Context(), parsed.String())
	if err != nil {
		if errors.Is(err, usecase.ErrProductNotFound) {
			api.SendNotFoundError(w, "product not found")
			return
		}

		slog.Error("failed to get product by id",
			slog.String("id", parsed.String()),
			slog.Any("error", err),
		)
		api.SendInternalError(w, "failed to get product by id")
		return
	}

	api.SendJSON(w, convertToProduct(product), http.StatusOK)
}

func convertToProduct(p models.Product) dto.Product {
	imageURLs := make([]string, 0, len(p.Images))
	for _, img := range p.Images {
		imageURLs = append(imageURLs, img.URL)
	}

	return dto.Product{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Price:       p.Price,
		ImageURLs:   imageURLs,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
