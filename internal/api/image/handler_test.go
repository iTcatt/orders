package image_test

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	imagehandler "iTcatt/orders/internal/api/image"
	"iTcatt/orders/internal/api/image/mocks"
	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
	ucimage "iTcatt/orders/internal/usecase/image"
)

const (
	validProductID = "01900000-0000-7000-8000-000000000064"
	validImageID   = "01900000-0000-7000-8000-000000000065"
	uuidV4         = "550e8400-e29b-41d4-a716-446655440000"
)

func buildUploadRequest(t *testing.T, productID, contentType string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"; filename="test.jpg"`)
	h.Set("Content-Type", contentType)
	part, err := mw.CreatePart(h)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	r := httptest.NewRequest(http.MethodPost, "/", &body)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.SetPathValue("id", productID)
	return r
}

func buildEmptyMultipartRequest(t *testing.T, productID string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	require.NoError(t, mw.Close())

	r := httptest.NewRequest(http.MethodPost, "/", &body)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.SetPathValue("id", productID)
	return r
}

func TestHandler_Upload(t *testing.T) {
	imgContent := []byte("fake-image-data")

	tests := []struct {
		name       string
		buildReq   func(t *testing.T) *http.Request
		setup      func(uc *mocks.MockimageUsecase)
		wantStatus int
		wantBody   string
	}{
		{
			name: "invalid product id",
			buildReq: func(t *testing.T) *http.Request {
				return buildUploadRequest(t, "not-a-uuid", "image/jpeg", imgContent)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "product id is uuid v4 not v7",
			buildReq: func(t *testing.T) *http.Request {
				return buildUploadRequest(t, uuidV4, "image/jpeg", imgContent)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing image field",
			buildReq: func(t *testing.T) *http.Request {
				return buildEmptyMultipartRequest(t, validProductID)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "unsupported mime type",
			buildReq: func(t *testing.T) *http.Request {
				return buildUploadRequest(t, validProductID, "image/gif", imgContent)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "product not found",
			buildReq: func(t *testing.T) *http.Request {
				return buildUploadRequest(t, validProductID, "image/jpeg", imgContent)
			},
			setup: func(uc *mocks.MockimageUsecase) {
				uc.EXPECT().Upload(mock.Anything, mock.Anything).Return(models.Image{}, usecase.ErrProductNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "too many images",
			buildReq: func(t *testing.T) *http.Request {
				return buildUploadRequest(t, validProductID, "image/jpeg", imgContent)
			},
			setup: func(uc *mocks.MockimageUsecase) {
				uc.EXPECT().Upload(mock.Anything, mock.Anything).Return(models.Image{}, ucimage.ErrTooManyImages).Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "internal error",
			buildReq: func(t *testing.T) *http.Request {
				return buildUploadRequest(t, validProductID, "image/jpeg", imgContent)
			},
			setup: func(uc *mocks.MockimageUsecase) {
				uc.EXPECT().Upload(mock.Anything, mock.Anything).Return(models.Image{}, errors.New("db error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			buildReq: func(t *testing.T) *http.Request {
				return buildUploadRequest(t, validProductID, "image/jpeg", imgContent)
			},
			setup: func(uc *mocks.MockimageUsecase) {
				uc.EXPECT().Upload(mock.Anything, mock.Anything).Return(models.Image{
					ID:  validImageID,
					URL: "http://minio.example.com/products/image.jpg",
				}, nil).Once()
			},
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"` + validImageID + `","url":"http://minio.example.com/products/image.jpg"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ucMock := mocks.NewMockimageUsecase(t)
			if tt.setup != nil {
				tt.setup(ucMock)
			}
			h := imagehandler.New(ucMock)
			w := httptest.NewRecorder()

			h.Upload(w, tt.buildReq(t))

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	buildReq := func(productID, imageID string) *http.Request {
		r := httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		r.SetPathValue("id", productID)
		r.SetPathValue("imageId", imageID)
		return r
	}

	tests := []struct {
		name       string
		productID  string
		imageID    string
		setup      func(uc *mocks.MockimageUsecase)
		wantStatus int
	}{
		{
			name:       "invalid product id",
			productID:  "not-a-uuid",
			imageID:    validImageID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "product id is uuid v4 not v7",
			productID:  uuidV4,
			imageID:    validImageID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid image id",
			productID:  validProductID,
			imageID:    "not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "image id is uuid v4 not v7",
			productID:  validProductID,
			imageID:    uuidV4,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "not found",
			productID: validProductID,
			imageID:   validImageID,
			setup: func(uc *mocks.MockimageUsecase) {
				uc.EXPECT().Delete(mock.Anything, validImageID).Return(usecase.ErrProductNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "internal error",
			productID: validProductID,
			imageID:   validImageID,
			setup: func(uc *mocks.MockimageUsecase) {
				uc.EXPECT().Delete(mock.Anything, validImageID).Return(errors.New("db error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "success",
			productID: validProductID,
			imageID:   validImageID,
			setup: func(uc *mocks.MockimageUsecase) {
				uc.EXPECT().Delete(mock.Anything, validImageID).Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ucMock := mocks.NewMockimageUsecase(t)
			if tt.setup != nil {
				tt.setup(ucMock)
			}
			h := imagehandler.New(ucMock)
			w := httptest.NewRecorder()

			h.Delete(w, buildReq(tt.productID, tt.imageID))

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
