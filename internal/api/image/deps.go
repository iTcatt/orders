package image

import (
	"context"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
)

type imageUsecase interface {
	Upload(ctx context.Context, input usecase.UploadImageIn) (models.Image, error)
	Delete(ctx context.Context, imageID string) error
}
