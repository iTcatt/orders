package category

import (
	"context"

	"iTcatt/orders/internal/models"
)

type categoryUsecase interface {
	GetAll(ctx context.Context) ([]models.Category, error)
}
