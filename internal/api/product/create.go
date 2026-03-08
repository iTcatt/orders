package product

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"iTcatt/orders/internal/api"
	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/usecase"
)

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	in, err := h.extractCreateInput(r)
	if err != nil {
		api.SendValidationError(w, err.Error())
		return
	}

	id, err := h.uc.CreateProduct(r.Context(), in)
	if err != nil {
		slog.Error("failed to create product", slog.String("error", err.Error()))
		api.SendInternalError(w, "failed to create product")
		return
	}

	out := dto.CreateProductOut{ID: id}
	api.SendJSON(w, out, http.StatusCreated)
}

func (h *handler) extractCreateInput(r *http.Request) (usecase.CreateProductIn, error) {
	var in dto.CreateProductIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return usecase.CreateProductIn{}, fmt.Errorf("invalid request body: %w", err)
	}

	if err := h.v.Struct(in); err != nil {
		return usecase.CreateProductIn{}, fmt.Errorf("validation: %w", err)
	}

	return usecase.CreateProductIn{
		Title:       in.Title,
		Description: in.Description,
		Price:       in.Price,
	}, nil
}
