package product

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/pkg/api"
)

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	idRaw, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil || idRaw == 0 {
		api.SendValidationError(w, "id must be positive")
		return
	}

	err = h.uc.DeleteProduct(r.Context(), uint32(idRaw))
	if err != nil {
		if errors.Is(err, usecase.ErrProductNotFound) {
			api.SendNotFoundError(w, "product not found")
			return
		}

		slog.Error("failed to delete product", slog.Any("error", err))
		api.SendInternalError(w, "failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
