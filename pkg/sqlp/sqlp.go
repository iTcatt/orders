package sqlp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)

type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}

func Get[T any](ctx context.Context, db *sqlx.DB, query sq.SelectBuilder) (T, error) {
	var result T

	q, args, err := query.ToSql()
	if err != nil {
		return result, err
	}

	err = dbFromCtx(ctx, db).GetContext(ctx, &result, q, args...)
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
		return nil, fmt.Errorf("build select query: %w", err)
	}

	var result []T
	err = dbFromCtx(ctx, db).SelectContext(ctx, &result, q, args...)
	if err != nil {
		return nil, fmt.Errorf("execute select query: %w", err)
	}

	return result, nil
}

func Insert(ctx context.Context, db *sqlx.DB, query sq.InsertBuilder) error {
	q, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = dbFromCtx(ctx, db).ExecContext(ctx, q, args...)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return fmt.Errorf("execute insert query: %w", err)
	}

	return nil
}

func Update(ctx context.Context, db *sqlx.DB, q sq.UpdateBuilder) error {
	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build update query: %w", err)
	}

	affected, err := dbFromCtx(ctx, db).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute update query: %w", err)
	}

	if rows, err := affected.RowsAffected(); err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	} else if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func Delete(ctx context.Context, db *sqlx.DB, q sq.DeleteBuilder) error {
	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}

	affected, err := dbFromCtx(ctx, db).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute delete query: %w", err)
	}

	if rows, err := affected.RowsAffected(); err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	} else if rows == 0 {
		return ErrNotFound
	}

	return nil
}
