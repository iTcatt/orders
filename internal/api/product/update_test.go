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
	in := dto.UpdateProductIn{
		Title:       new("Updated Product"),
		Description: new("Updated Description"),
		Price:       new(uint32(200)),
	}
	bytesIn, _ := json.Marshal(in)

	t.Run("success", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)

		d.uc.EXPECT().
			UpdateProduct(mock.Anything, productID, mock.Anything).
			Return(nil).
			Once()

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/product/%s", productID), bytes.NewBuffer(bytesIn))
		req.SetPathValue("id", productID)
		resp := httptest.NewRecorder()

		h.Update(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("not found", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)

		d.uc.EXPECT().
			UpdateProduct(mock.Anything, productID, mock.Anything).
			Return(usecase.ErrProductNotFound).
			Once()

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/product/%s", productID), bytes.NewBuffer(bytesIn))
		req.SetPathValue("id", productID)
		resp := httptest.NewRecorder()

		h.Update(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)

		d.uc.EXPECT().
			UpdateProduct(mock.Anything, productID, mock.Anything).
			Return(errors.New("some error")).
			Once()

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/product/%s", productID), bytes.NewBuffer(bytesIn))
		req.SetPathValue("id", productID)
		resp := httptest.NewRecorder()

		h.Update(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestHandler_UpdateValidation(t *testing.T) {
	h := product.New(nil)

	tests := []struct {
		name string
		in   string
		id   string
		want string
	}{
		{
			name: "negative price",
			in:   `{"title": "Test Product", "description": "Test Description", "price": -100}`,
			id:   productID,
			want: `{"message":"invalid input: json: cannot unmarshal number -100 into Go struct field UpdateProductIn.price of type uint32","code":400}`,
		},
		{
			name: "invalid JSON",
			id:   productID,
			in:   `not json`,
			want: `{"message":"invalid input: invalid character 'o' in literal null (expecting 'u')","code":400}`,
		},
		{
			name: "invalid id",
			id:   "not-a-uuid",
			in:   `{"title": "Test Product", "description": "Test Description", "price": 100}`,
			want: `{"message":"id must be a valid UUID v7","code":400}`,
		},
		{
			name: "empty payload",
			id:   productID,
			in:   `{}`,
			want: `{"message":"at least one field must be provided","code":400}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/product/%s", tt.id), bytes.NewBufferString(tt.in))
			r.SetPathValue("id", tt.id)

			h.Update(w, r)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, tt.want, strings.TrimSpace(w.Body.String()))
		})
	}
}
