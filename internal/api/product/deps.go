package product

import (
	"context"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
)

type productUsecase interface {
	GetProducts(ctx context.Context, in usecase.GetProductsIn) ([]models.Product, error)
	GetProductByID(ctx context.Context, id uint32) (models.Product, error)
	CreateProduct(ctx context.Context, in usecase.CreateProductIn) (uint32, error)
	UpdateProduct(ctx context.Context, id uint32, in usecase.UpdateProductIn) error
	DeleteProduct(ctx context.Context, id uint32) error
}
