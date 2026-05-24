package product

import (
	"context"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
)

type productUsecase interface {
	GetProducts(ctx context.Context, in usecase.GetProductsIn) ([]models.Product, error)
	GetProductByID(ctx context.Context, id string) (models.Product, error)
	CreateProduct(ctx context.Context, in usecase.CreateProductIn) (string, error)
	UpdateProduct(ctx context.Context, id string, in usecase.UpdateProductIn) error
	DeleteProduct(ctx context.Context, id string) error
}
