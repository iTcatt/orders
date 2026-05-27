package image_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	imagehandler "iTcatt/orders/internal/api/image"
	"iTcatt/orders/internal/usecase"
)

func TestHandler_Reorder(t *testing.T) {
	validImageID2 := "01900000-0000-7000-8000-000000000066"

	buildReq := func(productID, body string) *http.Request {
		r := httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(body))
		r.Header.Set("Content-Type", "application/json")
		r.SetPathValue("id", productID)
		return r
	}

	t.Run("success", func(t *testing.T) {
		d := setupDeps(t)
		d.uc.EXPECT().Reorder(mock.Anything, validProductID, []usecase.ImagePosition{
			{ID: validImageID2, Position: 0},
			{ID: validImageID, Position: 1},
		}).Return(nil).Once()

		w := httptest.NewRecorder()
		h := imagehandler.New(d.uc)
		h.Reorder(w, buildReq(validProductID,
			`[{"id":"`+validImageID2+`","position":0},{"id":"`+validImageID+`","position":1}]`))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		d := setupDeps(t)
		d.uc.EXPECT().Reorder(mock.Anything, validProductID, mock.Anything).Return(errors.New("db error")).Once()

		w := httptest.NewRecorder()
		h := imagehandler.New(d.uc)
		h.Reorder(w, buildReq(validProductID, `[{"id":"`+validImageID+`","position":0}]`))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid product id", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Reorder(w, buildReq("not-a-uuid", `[{"id":"`+validImageID+`","position":0}]`))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("product id is uuid v4 not v7", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Reorder(w, buildReq(uuidV4, `[{"id":"`+validImageID+`","position":0}]`))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Reorder(w, buildReq(validProductID, `not json`))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty array", func(t *testing.T) {
		h := imagehandler.New(nil)
		w := httptest.NewRecorder()
		h.Reorder(w, buildReq(validProductID, `[]`))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
