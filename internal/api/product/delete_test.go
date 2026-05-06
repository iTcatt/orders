package product_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"iTcatt/orders/internal/api/product"
	"iTcatt/orders/internal/usecase"
)

const productID = uint32(1)

func TestHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		d := setupDeps(t)
		handler := product.New(d.uc)

		d.uc.EXPECT().
			DeleteProduct(mock.Anything, productID).
			Return(nil).
			Once()

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/product/%d", productID), http.NoBody)
		req.SetPathValue("id", strconv.Itoa(int(productID)))
		resp := httptest.NewRecorder()

		handler.Delete(resp, req)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})

	t.Run("not found", func(t *testing.T) {
		d := setupDeps(t)
		handler := product.New(d.uc)

		d.uc.EXPECT().
			DeleteProduct(mock.Anything, productID).
			Return(usecase.ErrProductNotFound).
			Once()

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/product/%d", productID), http.NoBody)
		req.SetPathValue("id", strconv.Itoa(int(productID)))
		resp := httptest.NewRecorder()

		handler.Delete(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.JSONEq(t, `{"message":"product not found","code":404}`, resp.Body.String())
	})

	t.Run("internal error", func(t *testing.T) {
		d := setupDeps(t)
		handler := product.New(d.uc)

		d.uc.EXPECT().
			DeleteProduct(mock.Anything, productID).
			Return(errors.New("some error")).
			Once()

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/product/%d", productID), http.NoBody)
		req.SetPathValue("id", strconv.Itoa(int(productID)))
		resp := httptest.NewRecorder()

		handler.Delete(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.JSONEq(t, `{"message":"failed to delete product","code":500}`, resp.Body.String())
	})

	t.Run("validation error", func(t *testing.T) {
		handler := product.New(nil)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/product/%d", -100), http.NoBody)
		req.SetPathValue("id", strconv.Itoa(-100))
		resp := httptest.NewRecorder()

		handler.Delete(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.JSONEq(t, `{"message":"id must be positive","code":400}`, resp.Body.String())
	})
}
