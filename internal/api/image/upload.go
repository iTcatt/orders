package image

import (
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/http"

	"github.com/google/uuid"

	"iTcatt/orders/internal/api/image/dto"
	"iTcatt/orders/internal/usecase"
	ucimage "iTcatt/orders/internal/usecase/image"
	"iTcatt/orders/pkg/api"
)

const maxUploadSize = 5 << 20 // 5 MB

var allowedMIME = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

func (h *handler) Upload(w http.ResponseWriter, r *http.Request) {
	productID, err := parseProductID(r)
	if err != nil {
		api.SendValidationError(w, err.Error())
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1024)

	if err := validate(r); err != nil {
		api.SendValidationError(w, err.Error())
		return
	}

	input, err := extractInput(r, productID)
	if err != nil {
		api.SendValidationError(w, err.Error())
		return
	}
	defer input.File.Close()

	img, err := h.uc.Upload(r.Context(), input)
	if err != nil {
		if errors.Is(err, usecase.ErrProductNotFound) {
			api.SendNotFoundError(w, "product not found")
			return
		}
		if errors.Is(err, ucimage.ErrTooManyImages) {
			api.SendError(w, "product already has 3 images", http.StatusUnprocessableEntity)
			return
		}

		slog.Error("failed to upload image",
			slog.String("product_id", productID),
			slog.Any("error", err),
		)
		api.SendInternalError(w, "failed to upload image")
		return
	}

	api.SendJSON(w, dto.UploadResponse{ID: img.ID, URL: img.URL}, http.StatusCreated)
}

func validate(r *http.Request) error {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return errors.New("request too large or malformed multipart")
	}
	headers := r.MultipartForm.File["image"]
	if len(headers) == 0 {
		return errors.New("field 'image' is required")
	}
	header := headers[0]
	if header.Size > maxUploadSize {
		return fmt.Errorf("file too large, max %d MB", maxUploadSize>>20)
	}
	mediaType, _, _ := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if _, ok := allowedMIME[mediaType]; !ok {
		return errors.New("unsupported image type, allowed: jpeg, png, webp")
	}
	return nil
}

func extractInput(r *http.Request, productID string) (usecase.UploadImageIn, error) {
	file, header, err := r.FormFile("image")
	if err != nil {
		return usecase.UploadImageIn{}, errors.New("field 'image' is required")
	}
	mediaType, _, _ := mime.ParseMediaType(header.Header.Get("Content-Type"))
	return usecase.UploadImageIn{
		ProductID:   productID,
		File:        file,
		Size:        header.Size,
		ContentType: mediaType,
		Ext:         allowedMIME[mediaType],
	}, nil
}

func parseProductID(r *http.Request) (string, error) {
	parsed, err := uuid.Parse(r.PathValue("id"))
	if err != nil || parsed.Version() != 7 {
		return "", fmt.Errorf("id must be a valid UUID v7")
	}
	return parsed.String(), nil
}
