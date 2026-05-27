package products

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"iTcatt/orders/internal/models"
	st "iTcatt/orders/internal/storage"
	"iTcatt/orders/pkg/sqlp"
)

const productTable = "products"

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

func (s *storage) Get(ctx context.Context, in st.GetProductsIn) ([]models.Product, error) {
	query := s.builder.
		Select(getFields()...).
		From(productTable).
		OrderBy("created_at DESC").
		Limit(uint64(in.Limit)).
		Offset(uint64(in.Offset))

	if in.CategoryID != nil {
		query = query.Where(sq.Eq{"category_id": *in.CategoryID})
	}

	return sqlp.Select[models.Product](ctx, s.db, query)
}

func (s *storage) GetByID(ctx context.Context, id string) (models.Product, error) {
	query := s.builder.
		Select(getFields()...).
		From(productTable).
		Where(sq.Eq{"id": id})

	return sqlp.Get[models.Product](ctx, s.db, query)
}

func (s *storage) Create(ctx context.Context, p models.Product) error {
	query := s.builder.
		Insert(productTable).
		SetMap(p.ToMap())

	return sqlp.Insert(ctx, s.db, query)
}

func (s *storage) Delete(ctx context.Context, id string) error {
	query := s.builder.
		Delete(productTable).
		Where(sq.Eq{"id": id})

	return sqlp.Delete(ctx, s.db, query)
}

func (s *storage) Update(ctx context.Context, id string, in st.UpdateProductIn) error {
	query := s.builder.
		Update(productTable).
		SetMap(in.ToMap()).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id})

	return sqlp.Update(ctx, s.db, query)
}

func getFields() []string {
	return []string{
		"id",
		"title",
		"description",
		"price",
		"category_id",
		"created_at",
		"updated_at",
	}
}
