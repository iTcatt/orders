package image_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	imagehandler "iTcatt/orders/internal/api/image"
	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
	ucimage "iTcatt/orders/internal/usecase/image"
)

func TestHandler_Upload(t *testing.T) {
	imgContent := []byte("fake-image-data")

	t.Run("success", func(t *testing.T) {
		d := setupDeps(t)
		d.uc.EXPECT().Upload(mock.Anything, mock.Anything).Return(models.Image{
			ID:  validImageID,
			URL: "http://minio.example.com/products/image.jpg",
		}, nil).Once()

		w := httptest.NewRecorder()
		h := imagehandler.New(d.uc)
		h.Upload(w, buildUploadRequest(t, validProductID, "image/jpeg", imgContent))

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, `{"id":"`+validImageID+`","url":"http://minio.example.com/products/image.jpg"}`, w.Body.String())
	})

	t.Run("product not found", func(t *testing.T) {
		d := setupDeps(t)
		d.uc.EXPECT().Upload(mock.Anything, mock.Anything).Return(models.Image{}, usecase.ErrProductNotFound).Once()

		w := httptest.NewRecorder()
		h := imagehandler.New(d.uc)
		h.Upload(w, buildUploadRequest(t, validProductID, "image/jpeg", imgContent))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("too many images", func(t *testing.T) {
		d := setupDeps(t)
		d.uc.EXPECT().Upload(mock.Anything, mock.Anything).Return(models.Image{}, ucimage.ErrTooManyImages).Once()

		w := httptest.NewRecorder()
		h := imagehandler.New(d.uc)
		h.Upload(w, buildUploadRequest(t, validProductID, "image/jpeg", imgContent))

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		d := setupDeps(t)
		d.uc.EXPECT().Upload(mock.Anything, mock.Anything).Return(models.Image{}, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		h := imagehandler.New(d.uc)
		h.Upload(w, buildUploadRequest(t, validProductID, "image/jpeg", imgContent))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid product id", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Upload(w, buildUploadRequest(t, "not-a-uuid", "image/jpeg", imgContent))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("product id is uuid v4 not v7", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Upload(w, buildUploadRequest(t, uuidV4, "image/jpeg", imgContent))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing image field", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Upload(w, buildEmptyMultipartRequest(t, validProductID))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unsupported mime type", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Upload(w, buildUploadRequest(t, validProductID, "image/gif", imgContent))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
