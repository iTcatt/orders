package product_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"iTcatt/orders/internal/api/product"
	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
)

func TestHandler_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)

		products := []models.Product{
			{ID: 1, Title: "P1", Description: "Desc1", Price: 100},
			{ID: 2, Title: "P2", Description: "Desc2", Price: 200},
		}

		d.uc.EXPECT().
			GetProducts(mock.Anything, usecase.GetProductsIn{Page: 1, Limit: 10}).
			Return(products, nil).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/product/?page=1&limit=10", http.NoBody)
		resp := httptest.NewRecorder()

		h.Get(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), `"id":1`)
		assert.Contains(t, resp.Body.String(), `"id":2`)
	})

	t.Run("empty list", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)

		d.uc.EXPECT().
			GetProducts(mock.Anything, usecase.GetProductsIn{Page: 1, Limit: 10}).
			Return([]models.Product{}, nil).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/product/", http.NoBody)
		resp := httptest.NewRecorder()

		h.Get(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "[]\n", resp.Body.String())
	})

	t.Run("internal error", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)

		d.uc.EXPECT().
			GetProducts(mock.Anything, usecase.GetProductsIn{Page: 1, Limit: 10}).
			Return(nil, errors.New("some error")).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/product/", http.NoBody)
		resp := httptest.NewRecorder()

		h.Get(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Contains(t, resp.Body.String(), `"message":"failed to get products"`)
	})
}
