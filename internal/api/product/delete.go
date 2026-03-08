package product

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"iTcatt/orders/internal/api"
	"iTcatt/orders/internal/usecase/product"
)

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if id <= 0 {
		api.SendValidationError(w, "id must be positive")
		return
	}

	err := h.uc.DeleteProduct(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			api.SendNotFoundError(w, "product not found")
			return
		}

		slog.Error("failed to delete product", slog.String("error", err.Error()))
		api.SendInternalError(w, "failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
