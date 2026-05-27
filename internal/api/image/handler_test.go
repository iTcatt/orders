package image_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/require"

	"iTcatt/orders/internal/api/image/mocks"
)

const (
	validProductID = "01900000-0000-7000-8000-000000000064"
	validImageID   = "01900000-0000-7000-8000-000000000065"
	uuidV4         = "550e8400-e29b-41d4-a716-446655440000"
)

type deps struct {
	uc *mocks.MockimageUsecase
}

func setupDeps(t *testing.T) deps {
	t.Helper()
	return deps{
		uc: mocks.NewMockimageUsecase(t),
	}
}

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
