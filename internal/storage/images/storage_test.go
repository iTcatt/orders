package images_test

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
	"iTcatt/orders/internal/storage/images"
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

	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		panic("failed to set goose dialect: " + err.Error())
	}
	if err := goose.Up(testDB.DB, "../../../migrations/postgres/master"); err != nil {
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
		CategoryID:  1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func newImage(id, productID string, position int16) models.Image {
	return models.Image{
		ID:        id,
		ProductID: productID,
		URL:       "https://example.com/" + id + ".jpg",
		ObjectKey: "products/" + productID + "/" + id + ".jpg",
		Position:  position,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
}

func TestStorage_Create(t *testing.T) {
	truncate(t)
	ps := products.New(testDB)
	s := images.New(testDB)
	ctx := context.Background()

	productID := "01900000-0000-7000-8000-000000000001"
	require.NoError(t, ps.Create(ctx, newProduct(productID)))

	t.Run("creates image successfully", func(t *testing.T) {
		img := newImage("01900000-0000-7001-8000-000000000001", productID, 0)
		err := s.Create(ctx, img)
		require.NoError(t, err)

		got, err := s.GetByProductID(ctx, productID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, img.ID, got[0].ID)
		assert.Equal(t, img.URL, got[0].URL)
		assert.Equal(t, img.ObjectKey, got[0].ObjectKey)
		assert.Equal(t, img.Position, got[0].Position)
	})

	t.Run("returns ErrAlreadyExists on duplicate ID", func(t *testing.T) {
		img := newImage("01900000-0000-7001-8000-000000000001", productID, 1) // same ID
		err := s.Create(ctx, img)
		require.ErrorIs(t, err, sqlp.ErrAlreadyExists)
	})
}

func TestStorage_GetByProductID(t *testing.T) {
	truncate(t)
	ps := products.New(testDB)
	s := images.New(testDB)
	ctx := context.Background()

	t.Run("returns empty slice for unknown product", func(t *testing.T) {
		got, err := s.GetByProductID(ctx, "01900000-0000-7000-8000-999999999999")
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("returns images ordered by position", func(t *testing.T) {
		productID := "01900000-0000-7000-8000-000000000010"
		require.NoError(t, ps.Create(ctx, newProduct(productID)))

		img1 := newImage("01900000-0000-7001-8000-000000000011", productID, 1)
		img2 := newImage("01900000-0000-7001-8000-000000000012", productID, 0)
		require.NoError(t, s.Create(ctx, img1))
		require.NoError(t, s.Create(ctx, img2))

		got, err := s.GetByProductID(ctx, productID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, img2.ID, got[0].ID) // position 0 first
		assert.Equal(t, img1.ID, got[1].ID) // position 1 second
	})
}

func TestStorage_GetByProductIDs(t *testing.T) {
	truncate(t)
	ps := products.New(testDB)
	s := images.New(testDB)
	ctx := context.Background()

	t.Run("returns nil for empty input", func(t *testing.T) {
		got, err := s.GetByProductIDs(ctx, nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("returns empty for unknown product IDs", func(t *testing.T) {
		got, err := s.GetByProductIDs(ctx, []string{"01900000-0000-7000-8000-999999999998"})
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("returns images for multiple products", func(t *testing.T) {
		p1 := "01900000-0000-7000-8000-000000000020"
		p2 := "01900000-0000-7000-8000-000000000021"
		require.NoError(t, ps.Create(ctx, newProduct(p1)))
		require.NoError(t, ps.Create(ctx, newProduct(p2)))

		img1 := newImage("01900000-0000-7001-8000-000000000021", p1, 0)
		img2 := newImage("01900000-0000-7001-8000-000000000022", p2, 0)
		require.NoError(t, s.Create(ctx, img1))
		require.NoError(t, s.Create(ctx, img2))

		got, err := s.GetByProductIDs(ctx, []string{p1, p2})
		require.NoError(t, err)
		require.Len(t, got, 2)

		gotIDs := []string{got[0].ID, got[1].ID}
		assert.Contains(t, gotIDs, img1.ID)
		assert.Contains(t, gotIDs, img2.ID)
	})
}

func TestStorage_Count(t *testing.T) {
	truncate(t)
	ps := products.New(testDB)
	s := images.New(testDB)
	ctx := context.Background()

	t.Run("returns 0 for unknown product", func(t *testing.T) {
		count, err := s.Count(ctx, "01900000-0000-7000-8000-999999999999")
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("returns correct count", func(t *testing.T) {
		productID := "01900000-0000-7000-8000-000000000030"
		require.NoError(t, ps.Create(ctx, newProduct(productID)))

		imageIDs := []string{
			"01900000-0000-7001-8000-000000000031",
			"01900000-0000-7001-8000-000000000032",
			"01900000-0000-7001-8000-000000000033",
		}
		for i, id := range imageIDs {
			require.NoError(t, s.Create(ctx, newImage(id, productID, int16(i))))
		}

		count, err := s.Count(ctx, productID)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})
}

func TestStorage_Delete(t *testing.T) {
	truncate(t)
	ps := products.New(testDB)
	s := images.New(testDB)
	ctx := context.Background()

	t.Run("returns ErrNotFound for missing image", func(t *testing.T) {
		err := s.Delete(ctx, "01900000-0000-7001-8000-999999999999")
		require.ErrorIs(t, err, sqlp.ErrNotFound)
	})

	t.Run("deletes existing image", func(t *testing.T) {
		productID := "01900000-0000-7000-8000-000000000040"
		require.NoError(t, ps.Create(ctx, newProduct(productID)))

		img := newImage("01900000-0000-7001-8000-000000000041", productID, 0)
		require.NoError(t, s.Create(ctx, img))

		err := s.Delete(ctx, img.ID)
		require.NoError(t, err)

		count, err := s.Count(ctx, productID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
