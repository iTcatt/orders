package product

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/pkg/api"
)

func (h *handler) Update(w http.ResponseWriter, r *http.Request) {
	id, in, err := h.extractUpdateInput(r)
	if err != nil {
		api.SendValidationError(w, err.Error())
		return
	}

	if err := h.uc.UpdateProduct(r.Context(), id, in); err != nil {
		if errors.Is(err, usecase.ErrProductNotFound) {
			api.SendNotFoundError(w, "product not found")
			return
		}

		slog.Error("failed to update product",
			slog.String("id", id),
			slog.Any("error", err),
		)
		api.SendInternalError(w, "failed to update product")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) extractUpdateInput(r *http.Request) (string, usecase.UpdateProductIn, error) {
	defer r.Body.Close()

	parsed, err := uuid.Parse(r.PathValue("id"))
	if err != nil || parsed.Version() != 7 {
		return "", usecase.UpdateProductIn{}, fmt.Errorf("id must be a valid UUID v7")
	}

	var in dto.UpdateProductIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return "", usecase.UpdateProductIn{}, fmt.Errorf("invalid input: %w", err)
	}

	if err := h.v.Struct(in); err != nil {
		return "", usecase.UpdateProductIn{}, fmt.Errorf("validation: %w", err)
	}

	if in.Title == nil && in.Description == nil && in.Price == nil {
		return "", usecase.UpdateProductIn{}, fmt.Errorf("at least one field must be provided")
	}

	return parsed.String(), usecase.UpdateProductIn{
		Title:       in.Title,
		Description: in.Description,
		Price:       in.Price,
	}, nil
}
