package images

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/pkg/sqlp"
)

const imageTable = "product_images"

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

func (s *storage) Create(ctx context.Context, img models.Image) error {
	query := s.builder.
		Insert(imageTable).
		Columns("id", "product_id", "url", "object_key", "position", "created_at").
		Values(img.ID, img.ProductID, img.URL, img.ObjectKey, img.Position, img.CreatedAt)

	return sqlp.Insert(ctx, s.db, query)
}

func (s *storage) GetByProductIDs(ctx context.Context, productIDs []string) ([]models.Image, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	ids := make([]any, len(productIDs))
	for i, id := range productIDs {
		ids[i] = id
	}

	query := s.builder.
		Select("id", "product_id", "url", "object_key", "position", "created_at").
		From(imageTable).
		Where(sq.Eq{"product_id": ids}).
		OrderBy("product_id", "position ASC")

	return sqlp.Select[models.Image](ctx, s.db, query)
}

func (s *storage) Get(ctx context.Context, imageID string) (models.Image, error) {
	query := s.builder.
		Select("id", "product_id", "url", "object_key", "position", "created_at").
		From(imageTable).
		Where(sq.Eq{"id": imageID})

	return sqlp.Get[models.Image](ctx, s.db, query)
}

func (s *storage) GetByProductID(ctx context.Context, productID string) ([]models.Image, error) {
	query := s.builder.
		Select("id", "product_id", "url", "object_key", "position", "created_at").
		From(imageTable).
		Where(sq.Eq{"product_id": productID}).
		OrderBy("position ASC")

	return sqlp.Select[models.Image](ctx, s.db, query)
}

func (s *storage) Count(ctx context.Context, productID string) (int, error) {
	query := s.builder.
		Select("COUNT(*)").
		From(imageTable).
		Where(sq.Eq{"product_id": productID})

	q, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count query: %w", err)
	}

	var count int
	if err := s.db.GetContext(ctx, &count, q, args...); err != nil {
		return 0, fmt.Errorf("execute count query: %w", err)
	}

	return count, nil
}

func (s *storage) Delete(ctx context.Context, imageID string) error {
	query := s.builder.
		Delete(imageTable).
		Where(sq.Eq{"id": imageID})

	return sqlp.Delete(ctx, s.db, query)
}
