package product

import (
	"context"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
)

type productUsecase interface {
	GetProductByID(ctx context.Context, id int32) (models.Product, error)
	CreateProduct(ctx context.Context, in usecase.CreateProductIn) (int32, error)
	UpdateProduct(ctx context.Context, id int32, in usecase.UpdateProductIn) error
	DeleteProduct(ctx context.Context, id int32) error
}
