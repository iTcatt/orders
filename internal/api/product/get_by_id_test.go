package product_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"iTcatt/orders/internal/api/product"
	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
)

func TestHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)
		mux := http.NewServeMux()
		mux.HandleFunc("GET /product/{id}", h.GetByID)

		id := uint32(1)
		p := models.Product{
			ID:          id,
			Title:       "Test Product",
			Description: "Test Description",
			Price:       100,
		}

		d.uc.EXPECT().
			GetProductByID(mock.Anything, id).
			Return(p, nil).
			Once()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/product/%d", id), http.NoBody)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), `"id":1`)
		assert.Contains(t, resp.Body.String(), `"title":"Test Product"`)
	})

	t.Run("not found", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)
		mux := http.NewServeMux()
		mux.HandleFunc("GET /product/{id}", h.GetByID)

		id := uint32(1)

		d.uc.EXPECT().
			GetProductByID(mock.Anything, id).
			Return(models.Product{}, usecase.ErrProductNotFound).
			Once()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/product/%d", id), http.NoBody)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Contains(t, resp.Body.String(), `"message":"product not found"`)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := product.New(nil)
		mux := http.NewServeMux()
		mux.HandleFunc("GET /product/{id}", h.GetByID)

		req := httptest.NewRequest(http.MethodGet, "/product/-100", http.NoBody)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), `"message":"id must be positive"`)
	})

	t.Run("internal error", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)
		mux := http.NewServeMux()
		mux.HandleFunc("GET /product/{id}", h.GetByID)

		id := uint32(1)

		d.uc.EXPECT().
			GetProductByID(mock.Anything, id).
			Return(models.Product{}, errors.New("some error")).
			Once()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/product/%d", id), http.NoBody)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Contains(t, resp.Body.String(), `"message":"failed to get product by id"`)
	})
}
