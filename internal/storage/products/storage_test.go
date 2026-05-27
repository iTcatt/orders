package products_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"iTcatt/orders/internal/models"
	st "iTcatt/orders/internal/storage"
	"iTcatt/orders/internal/storage/products"
	"iTcatt/orders/pkg/sqlp"
)

var testDB *sqlx.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.WithSQLDriver("pgx"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer pgContainer.Terminate(ctx) //nolint:errcheck

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}

	testDB, err = sqlx.Open("pgx", connStr)
	if err != nil {
		panic("failed to open db: " + err.Error())
	}
	defer testDB.Close() //nolint:errcheck

	if err := testDB.Ping(); err != nil {
		panic("failed to ping db: " + err.Error())
	}

	migrationsDir := "../../../migrations/postgres/master"

	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		panic("failed to set goose dialect: " + err.Error())
	}
	if err := goose.Up(testDB.DB, migrationsDir); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	m.Run()
}

func truncate(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec("TRUNCATE TABLE products CASCADE")
	require.NoError(t, err)
}

func newProduct(id string) models.Product {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return models.Product{
		ID:          id,
		Title:       "Test Product",
		Description: "Test Description",
		Price:       9900,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestStorage_Create(t *testing.T) {
	truncate(t)
	s := products.New(testDB)
	ctx := context.Background()

	t.Run("creates product successfully", func(t *testing.T) {
		p := newProduct("01900000-0000-7000-8000-000000000001")
		err := s.Create(ctx, p)
		require.NoError(t, err)

		got, err := s.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, p.ID, got.ID)
		assert.Equal(t, p.Title, got.Title)
		assert.Equal(t, p.Description, got.Description)
		assert.Equal(t, p.Price, got.Price)
	})

	t.Run("returns ErrAlreadyExists on duplicate ID", func(t *testing.T) {
		p := newProduct("01900000-0000-7000-8000-000000000001") // same ID as above
		err := s.Create(ctx, p)
		require.ErrorIs(t, err, sqlp.ErrAlreadyExists)
	})
}

func TestStorage_GetByID(t *testing.T) {
	truncate(t)
	s := products.New(testDB)
	ctx := context.Background()

	t.Run("returns ErrNotFound for missing product", func(t *testing.T) {
		_, err := s.GetByID(ctx, "01900000-0000-7000-8000-999999999999")
		require.ErrorIs(t, err, sqlp.ErrNotFound)
	})

	t.Run("returns product by ID", func(t *testing.T) {
		p := newProduct("01900000-0000-7000-8000-000000000002")
		require.NoError(t, s.Create(ctx, p))

		got, err := s.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, p.ID, got.ID)
		assert.Equal(t, p.Title, got.Title)
	})
}

func TestStorage_Get(t *testing.T) {
	truncate(t)
	s := products.New(testDB)
	ctx := context.Background()

	ids := []string{
		"01900000-0000-7000-8000-000000000011",
		"01900000-0000-7000-8000-000000000012",
		"01900000-0000-7000-8000-000000000013",
		"01900000-0000-7000-8000-000000000014",
		"01900000-0000-7000-8000-000000000015",
	}
	for _, id := range ids {
		p := newProduct(id)
		p.Title = "Product"
		require.NoError(t, s.Create(ctx, p))
	}

	t.Run("returns all products with high limit", func(t *testing.T) {
		got, err := s.Get(ctx, st.GetProductsIn{Limit: 10, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, got, 5)
	})

	t.Run("respects limit", func(t *testing.T) {
		got, err := s.Get(ctx, st.GetProductsIn{Limit: 2, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("respects offset", func(t *testing.T) {
		got, err := s.Get(ctx, st.GetProductsIn{Limit: 10, Offset: 3})
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("returns empty slice when no products match", func(t *testing.T) {
		got, err := s.Get(ctx, st.GetProductsIn{Limit: 10, Offset: 100})
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestStorage_Update(t *testing.T) {
	truncate(t)
	s := products.New(testDB)
	ctx := context.Background()

	t.Run("returns ErrNotFound for missing product", func(t *testing.T) {
		title := "New Title"
		err := s.Update(ctx, "01900000-0000-7000-8000-999999999999", st.UpdateProductIn{Title: &title})
		require.ErrorIs(t, err, sqlp.ErrNotFound)
	})

	t.Run("updates title", func(t *testing.T) {
		p := newProduct("01900000-0000-7000-8000-000000000010")
		require.NoError(t, s.Create(ctx, p))

		newTitle := "Updated Title"
		err := s.Update(ctx, p.ID, st.UpdateProductIn{Title: &newTitle})
		require.NoError(t, err)

		got, err := s.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, newTitle, got.Title)
		assert.Equal(t, p.Description, got.Description)
		assert.Equal(t, p.Price, got.Price)
	})

	t.Run("updates multiple fields", func(t *testing.T) {
		p := newProduct("01900000-0000-7000-8000-000000000011")
		require.NoError(t, s.Create(ctx, p))

		newTitle := "Multi Update"
		newDesc := "New Desc"
		newPrice := uint32(5000)
		err := s.Update(ctx, p.ID, st.UpdateProductIn{
			Title:       &newTitle,
			Description: &newDesc,
			Price:       &newPrice,
		})
		require.NoError(t, err)

		got, err := s.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, newTitle, got.Title)
		assert.Equal(t, newDesc, got.Description)
		assert.Equal(t, newPrice, got.Price)
	})
}

func TestStorage_Delete(t *testing.T) {
	truncate(t)
	s := products.New(testDB)
	ctx := context.Background()

	t.Run("returns ErrNotFound for missing product", func(t *testing.T) {
		err := s.Delete(ctx, "01900000-0000-7000-8000-999999999999")
		require.ErrorIs(t, err, sqlp.ErrNotFound)
	})

	t.Run("deletes existing product", func(t *testing.T) {
		p := newProduct("01900000-0000-7000-8000-000000000020")
		require.NoError(t, s.Create(ctx, p))

		err := s.Delete(ctx, p.ID)
		require.NoError(t, err)

		_, err = s.GetByID(ctx, p.ID)
		require.ErrorIs(t, err, sqlp.ErrNotFound)
	})
}
