package image

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"iTcatt/orders/internal/api/image/dto"
	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/pkg/api"
)

func (h *handler) Reorder(w http.ResponseWriter, r *http.Request) {
	productID, err := parseProductID(r)
	if err != nil {
		api.SendValidationError(w, err.Error())
		return
	}

	positions, err := extractReorderInput(r)
	if err != nil {
		api.SendValidationError(w, err.Error())
		return
	}

	if err := h.uc.Reorder(r.Context(), productID, positions); err != nil {
		slog.Error("failed to reorder images",
			slog.String("product_id", productID),
			slog.Any("error", err),
		)
		api.SendInternalError(w, "failed to reorder images")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func extractReorderInput(r *http.Request) ([]usecase.ImagePosition, error) {
	defer r.Body.Close()

	var in []dto.ImagePositionIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if len(in) == 0 {
		return nil, fmt.Errorf("image_positions must not be empty")
	}

	positions := make([]usecase.ImagePosition, len(in))
	for i, p := range in {
		if p.ID == "" {
			return nil, fmt.Errorf("image id must not be empty")
		}
		positions[i] = usecase.ImagePosition{ID: p.ID, Position: p.Position}
	}

	return positions, nil
}
