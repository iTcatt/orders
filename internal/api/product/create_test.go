package product_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"iTcatt/orders/internal/api/product"
	"iTcatt/orders/internal/api/product/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Create(t *testing.T) {
	in := dto.CreateProductIn{
		Title:       "Test Product",
		Description: "Test Description",
		Price:       100,
	}
	bytesIn, _ := json.Marshal(in)

	t.Run("success", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)
		id := uint32(42)

		d.uc.EXPECT().
			CreateProduct(mock.Anything, mock.Anything).
			Return(id, nil).
			Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/product/", bytes.NewBuffer(bytesIn))

		h.Create(w, r)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, `{"id":42}`, strings.TrimSpace(w.Body.String()))
	})

	t.Run("usecase error", func(t *testing.T) {
		d := setupDeps(t)
		h := product.New(d.uc)

		d.uc.EXPECT().
			CreateProduct(mock.Anything, mock.Anything).
			Return(0, errors.New("some error")).
			Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/product/", bytes.NewBuffer(bytesIn))

		h.Create(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, `{"message":"failed to create product","code":500}`, strings.TrimSpace(w.Body.String()))
	})
}

func TestHandler_CreateValidation(t *testing.T) {
	h := product.New(nil)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty title",
			in:   `{"description": "desc", "price": 100}`,
			want: `{"message":"validation: Key: 'CreateProductIn.Title' Error:Field validation for 'Title' failed on the 'required' tag","code":400}`,
		},
		{
			name: "empty description",
			in:   `{"title": "Test Product", "price": 100}`,
			want: `{"message":"validation: Key: 'CreateProductIn.Description' Error:Field validation for 'Description' failed on the 'required' tag","code":400}`,
		},
		{
			name: "empty price",
			in:   `{"title": "Test Product", "description": "Test Description"}`,
			want: `{"message":"validation: Key: 'CreateProductIn.Price' Error:Field validation for 'Price' failed on the 'required' tag","code":400}`,
		},
		{
			name: "negative price",
			in:   `{"title": "Test Product", "description": "Test Description", "price": -100}`,
			want: `{"message":"invalid request body: json: cannot unmarshal number -100 into Go struct field CreateProductIn.price of type uint32","code":400}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/product/", bytes.NewBufferString(tt.in))

			h.Create(w, r)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, tt.want, strings.TrimSpace(w.Body.String()))
		})
	}
}
