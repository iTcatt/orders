package sqlp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)

func Get[T any](ctx context.Context, db *sqlx.DB, query sq.SelectBuilder) (T, error) {
	var result T

	q, args, err := query.ToSql()
	if err != nil {
		return result, err
	}

	err = db.GetContext(ctx, &result, q, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, ErrNotFound
		}
		return result, err
	}

	return result, nil
}

func Select[T any](ctx context.Context, db *sqlx.DB, query sq.SelectBuilder) ([]T, error) {
	q, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var result []T
	err = db.SelectContext(ctx, &result, q, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func Insert[T any](ctx context.Context, db *sqlx.DB, query sq.InsertBuilder) error {
	q, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	res, err := db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("failed to execute insert query: %w", err)
	}

	if rows, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	} else if rows == 0 {
		return ErrAlreadyExists
	}

	return nil
}

func Update[T any](ctx context.Context, db *sqlx.DB, q sq.UpdateBuilder) error {
	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	affected, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	if rows, err := affected.RowsAffected(); err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	} else if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func Delete[T any](ctx context.Context, db *sqlx.DB, q sq.DeleteBuilder) error {
	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	affected, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}

	if rows, err := affected.RowsAffected(); err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	} else if rows == 0 {
		return ErrNotFound
	}

	return nil
}
