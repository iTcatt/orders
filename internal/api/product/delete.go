package product

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/pkg/api"
)

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	parsed, err := uuid.Parse(r.PathValue("id"))
	if err != nil || parsed.Version() != 7 {
		api.SendValidationError(w, "id must be a valid UUID v7")
		return
	}

	err = h.uc.DeleteProduct(r.Context(), parsed.String())
	if err != nil {
		if errors.Is(err, usecase.ErrProductNotFound) {
			api.SendNotFoundError(w, "product not found")
			return
		}

		slog.Error("failed to delete product",
			slog.String("id", parsed.String()),
			slog.Any("error", err),
		)
		api.SendInternalError(w, "failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
