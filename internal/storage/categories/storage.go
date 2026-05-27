package categories

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/pkg/sqlp"
)

const categoryTable = "categories"

type storage struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
}

func New(db *sqlx.DB) *storage {
	return &storage{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *storage) GetAll(ctx context.Context) ([]models.Category, error) {
	query := s.builder.
		Select("id", "slug", "name").
		From(categoryTable).
		OrderBy("id ASC")

	return sqlp.Select[models.Category](ctx, s.db, query)
}
