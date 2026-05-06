package product

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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

		slog.Error("failed to update product", slog.Any("error", err))
		api.SendInternalError(w, "failed to update product")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) extractUpdateInput(r *http.Request) (uint32, usecase.UpdateProductIn, error) {
	defer r.Body.Close()

	idRaw, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil || idRaw == 0 {
		return 0, usecase.UpdateProductIn{}, fmt.Errorf("id must be positive")
	}

	var in dto.UpdateProductIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return 0, usecase.UpdateProductIn{}, fmt.Errorf("invalid input: %w", err)
	}

	if err := h.v.Struct(in); err != nil {
		return 0, usecase.UpdateProductIn{}, fmt.Errorf("validation: %w", err)
	}

	if in.Title == nil && in.Description == nil && in.Price == nil {
		return 0, usecase.UpdateProductIn{}, fmt.Errorf("at least one field must be provided")
	}

	return uint32(idRaw), usecase.UpdateProductIn{
		Title:       in.Title,
		Description: in.Description,
		Price:       in.Price,
	}, nil
}
