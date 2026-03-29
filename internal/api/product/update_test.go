package product_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"iTcatt/orders/internal/api/product"
	"iTcatt/orders/internal/api/product/dto"
	"iTcatt/orders/internal/usecase"
)

func TestHandler_Update(t *testing.T) {
	productID := uint32(1)
	in := dto.UpdateProductIn{
		Title:       new("Updated Product"),
		Description: new("Updated Description"),
		Price:       new(uint32(200)),
	}
	bytesIn, _ := json.Marshal(in)

	t.Run("success", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /product/{id}", h.Update)

		d.uc.EXPECT().
			UpdateProduct(mock.Anything, productID, mock.Anything).
			Return(nil).
			Once()

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/product/%d", productID), bytes.NewBuffer(bytesIn))
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("not found", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /product/{id}", h.Update)

		d.uc.EXPECT().
			UpdateProduct(mock.Anything, productID, mock.Anything).
			Return(usecase.ErrProductNotFound).
			Once()

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/product/%d", productID), bytes.NewBuffer(bytesIn))
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /product/{id}", h.Update)

		d.uc.EXPECT().
			UpdateProduct(mock.Anything, productID, mock.Anything).
			Return(errors.New("some error")).
			Once()

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/product/%d", productID), bytes.NewBuffer(bytesIn))
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestHandler_UpdateValidation(t *testing.T) {
	h := product.New(nil)
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /product/{id}", h.Update)

	tests := []struct {
		name string
		in   string
		id   int
		want string
	}{
		{
			name: "negative price",
			in:   `{"title": "Test Product", "description": "Test Description", "price": -100}`,
			id:   1,
			want: `{"message":"invalid input: json: cannot unmarshal number -100 into Go struct field UpdateProductIn.price of type uint32","code":400}`,
		},
		{
			name: "invalid JSON",
			id:   1,
			in:   `not json`,
			want: `{"message":"invalid input: invalid character 'o' in literal null (expecting 'u')","code":400}`,
		},
		{
			name: "invalid id",
			id:   0,
			in:   `{"title": "Test Product", "description": "Test Description", "price": 100}`,
			want: `{"message":"id must be positive","code":400}`,
		},
		{
			name: "empty payload",
			id:   1,
			in:   `{}`,
			want: `{"message":"at least one field must be provided","code":400}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/product/%d", tt.id), bytes.NewBufferString(tt.in))

			mux.ServeHTTP(w, r)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, tt.want, strings.TrimSpace(w.Body.String()))
		})
	}
}
