package image_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	imagehandler "iTcatt/orders/internal/api/image"
	"iTcatt/orders/internal/usecase"
)

func TestHandler_Delete(t *testing.T) {
	buildReq := func(productID, imageID string) *http.Request {
		r := httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		r.SetPathValue("id", productID)
		r.SetPathValue("imageId", imageID)
		return r
	}

	t.Run("success", func(t *testing.T) {
		d := setupDeps(t)
		d.uc.EXPECT().Delete(mock.Anything, validImageID).Return(nil).Once()

		w := httptest.NewRecorder()
		h := imagehandler.New(d.uc)
		h.Delete(w, buildReq(validProductID, validImageID))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		d := setupDeps(t)
		d.uc.EXPECT().Delete(mock.Anything, validImageID).Return(usecase.ErrProductNotFound).Once()

		w := httptest.NewRecorder()
		h := imagehandler.New(d.uc)
		h.Delete(w, buildReq(validProductID, validImageID))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"message":"image not found","code":404}`, w.Body.String())
	})

	t.Run("internal error", func(t *testing.T) {
		d := setupDeps(t)
		d.uc.EXPECT().Delete(mock.Anything, validImageID).Return(errors.New("db error")).Once()

		w := httptest.NewRecorder()
		h := imagehandler.New(d.uc)
		h.Delete(w, buildReq(validProductID, validImageID))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid product id", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Delete(w, buildReq("not-a-uuid", validImageID))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("product id is uuid v4 not v7", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Delete(w, buildReq(uuidV4, validImageID))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid image id", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Delete(w, buildReq(validProductID, "not-a-uuid"))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("image id is uuid v4 not v7", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Delete(w, buildReq(validProductID, uuidV4))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
