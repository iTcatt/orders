package category

import (
	"context"
	"fmt"

	"iTcatt/orders/internal/models"
)

type usecase struct {
	repo categoryRepo
}

func New(repo categoryRepo) *usecase {
	return &usecase{repo: repo}
}

func (u *usecase) GetAll(ctx context.Context) ([]models.Category, error) {
	categories, err := u.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}
	return categories, nil
}
