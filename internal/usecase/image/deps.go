package image

import (
	"context"
	"io"

	"iTcatt/orders/internal/models"
)

type imageRepo interface {
	Create(ctx context.Context, img models.Image) error
	Get(ctx context.Context, imageID string) (models.Image, error)
	GetByProductID(ctx context.Context, productID string) ([]models.Image, error)
	Count(ctx context.Context, productID string) (int, error)
	Delete(ctx context.Context, imageID string) error
}

type objectStore interface {
	URL(key string) string
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
}

type productRepo interface {
	GetByID(ctx context.Context, id string) (models.Product, error)
}

type txManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
