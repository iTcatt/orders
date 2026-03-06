package products

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/storage"
	"iTcatt/orders/pkg/sqlp"
)

const productTable = "products"

type Storage struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
}

func New(db *sqlx.DB) *Storage {
	return &Storage{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *Storage) GetByID(ctx context.Context, id int32) (models.Product, error) {
	query := s.builder.
		Select(getFields()...).
		From(productTable).
		Where(sq.Eq{"id": id})

	return sqlp.Get[models.Product](ctx, s.db, query)
}

func (s *Storage) Create(ctx context.Context, p models.Product) error {
	query := s.builder.
		Insert(productTable).
		SetMap(p.ToMap())

	return sqlp.Insert[models.Product](ctx, s.db, query)
}

func (s *Storage) Delete(ctx context.Context, id int32) error {
	query := s.builder.
		Delete(productTable).
		Where(sq.Eq{"id": id})

	return sqlp.Delete[models.Product](ctx, s.db, query)
}

func (s *Storage) Update(ctx context.Context, id int32, in storage.UpdateProductIn) error {
	query := s.builder.
		Update(productTable).
		SetMap(in.ToMap()).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id})

	return sqlp.Update[models.Product](ctx, s.db, query)
}

func getFields() []string {
	return []string{
		"id",
		"title",
		"description",
		"price",
		"created_at",
		"updated_at",
	}
}
