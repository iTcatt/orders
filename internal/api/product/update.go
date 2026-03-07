package product

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"iTcatt/orders/internal/storage"
	"iTcatt/orders/pkg/sqlp"
)

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, in, err := extractInput(r)
	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	if err := h.storage.Update(r.Context(), id, in); err != nil {
		if errors.Is(err, sqlp.ErrNotFound) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func extractInput(r *http.Request) (int32, storage.UpdateProductIn, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, storage.UpdateProductIn{}, fmt.Errorf("invalid id: %w", err)
	}

	var in storage.UpdateProductIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return 0, storage.UpdateProductIn{}, fmt.Errorf("invalid input: %w", err)
	}

	return int32(id), in, nil
}
