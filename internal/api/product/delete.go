package product

import (
	"errors"
	"net/http"
	"strconv"

	"iTcatt/orders/internal/usecase/product"
)

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.uc.DeleteProduct(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to delete product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
