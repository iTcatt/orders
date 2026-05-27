package image

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/pkg/api"
)

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	productID, err := parseProductID(r)
	if err != nil {
		api.SendValidationError(w, err.Error())
		return
	}

	imageID, err := parseImageID(r)
	if err != nil {
		api.SendValidationError(w, err.Error())
		return
	}

	if err := h.uc.Delete(r.Context(), imageID); err != nil {
		if errors.Is(err, usecase.ErrProductNotFound) {
			api.SendNotFoundError(w, "image not found")
			return
		}

		slog.Error("failed to delete image",
			slog.String("product_id", productID),
			slog.String("image_id", imageID),
			slog.Any("error", err),
		)
		api.SendInternalError(w, "failed to delete image")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseImageID(r *http.Request) (string, error) {
	parsed, err := uuid.Parse(r.PathValue("imageId"))
	if err != nil || parsed.Version() != 7 {
		return "", fmt.Errorf("imageId must be a valid UUID v7")
	}
	return parsed.String(), nil
}
