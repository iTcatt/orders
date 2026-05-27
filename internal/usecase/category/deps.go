package category

import (
	"context"

	"iTcatt/orders/internal/models"
)

type categoryRepo interface {
	GetAll(ctx context.Context) ([]models.Category, error)
}
