package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/storage"
	uc "iTcatt/orders/internal/usecase"
	"iTcatt/orders/pkg/sqlp"
)

type usecase struct {
	repo        productRepo
	now         func() time.Time
	idGenerator func() string
}

func New(
	repo productRepo,
	now func() time.Time,
	idGen func() string,
) *usecase {
	return &usecase{
		repo:        repo,
		now:         now,
		idGenerator: idGen,
	}
}

func (u *usecase) GetProducts(ctx context.Context, in uc.GetProductsIn) ([]models.Product, error) {
	products, err := u.repo.Get(ctx, storage.GetProductsIn{
		Limit:  in.Limit,
		Offset: (in.Page - 1) * in.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}

	return products, nil
}

func (u *usecase) GetProductByID(ctx context.Context, id string) (models.Product, error) {
	product, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sqlp.ErrNotFound) {
			return models.Product{}, uc.ErrProductNotFound
		}

		return models.Product{}, fmt.Errorf("get product by id: %w", err)
	}

	return product, nil
}

func (u *usecase) CreateProduct(ctx context.Context, in uc.CreateProductIn) (string, error) {
	now := u.now()
	product := models.Product{
		ID:          u.idGenerator(),
		Title:       in.Title,
		Description: in.Description,
		Price:       in.Price,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := u.repo.Create(ctx, product); err != nil {
		return "", fmt.Errorf("create product: %w", err)
	}

	return product.ID, nil
}

func (u *usecase) UpdateProduct(ctx context.Context, id string, in uc.UpdateProductIn) error {
	err := u.repo.Update(ctx, id, storage.UpdateProductIn{
		Title:       in.Title,
		Description: in.Description,
		Price:       in.Price,
	})
	if err != nil {
		if errors.Is(err, sqlp.ErrNotFound) {
			return uc.ErrProductNotFound
		}

		return fmt.Errorf("update product: %w", err)
	}

	return nil
}

func (u *usecase) DeleteProduct(ctx context.Context, id string) error {
	err := u.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sqlp.ErrNotFound) {
			return uc.ErrProductNotFound
		}

		return fmt.Errorf("delete product: %w", err)
	}

	return nil
}
