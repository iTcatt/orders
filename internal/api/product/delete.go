package product

import (
	"errors"
	"net/http"
	"strconv"

	"iTcatt/orders/pkg/sqlp"
)

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.storage.Delete(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sqlp.ErrNotFound) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		
		http.Error(w, "failed to delete product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
