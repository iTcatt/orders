package product_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"iTcatt/orders/internal/api/product"
	"iTcatt/orders/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const productID = uint32(1)

func TestHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		d := setupDeps(t)
		handler := product.New(d.uc)
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /product/{id}", handler.Delete)

		d.uc.EXPECT().
			DeleteProduct(mock.Anything, productID).
			Return(nil).
			Once()

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/product/%d", productID), nil)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})

	t.Run("not found", func(t *testing.T) {
		d := setupDeps(t)
		handler := product.New(d.uc)
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /product/{id}", handler.Delete)

		d.uc.EXPECT().
			DeleteProduct(mock.Anything, productID).
			Return(usecase.ErrProductNotFound).
			Once()

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/product/%d", productID), nil)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, `{"message":"product not found","code":404}`, strings.TrimSpace(resp.Body.String()))
	})

	t.Run("internal error", func(t *testing.T) {
		d := setupDeps(t)
		handler := product.New(d.uc)
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /product/{id}", handler.Delete)

		d.uc.EXPECT().
			DeleteProduct(mock.Anything, productID).
			Return(errors.New("some error")).
			Once()

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/product/%d", productID), nil)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, `{"message":"failed to delete product","code":500}`, strings.TrimSpace(resp.Body.String()))
	})
	
	t.Run("validation error", func(t *testing.T) {
		handler := product.New(nil)
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /product/{id}", handler.Delete)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/product/%d", -100), nil)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, `{"message":"id must be positive","code":400}`, strings.TrimSpace(resp.Body.String()))
	})
}