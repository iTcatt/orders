package product

import (
	"context"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/storage"
)

type productRepo interface {
	GetByID(ctx context.Context, id int32) (models.Product, error)
	Create(ctx context.Context, product models.Product) error
	Update(ctx context.Context, id int32, in storage.UpdateProductIn) error
	Delete(ctx context.Context, id int32) error
}
