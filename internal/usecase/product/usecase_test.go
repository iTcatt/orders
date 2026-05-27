package product_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/storage"
	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/internal/usecase/product"
	"iTcatt/orders/internal/usecase/product/mocks"
	"iTcatt/orders/pkg/sqlp"
)

const productID = "01900000-0000-7000-8000-000000000064"

type deps struct {
	productRepo *mocks.MockproductRepo
	imageRepo   *mocks.MockimageRepo
	now         func() time.Time
	idGenerator func() string
}

func setupDeps(t *testing.T) *deps {
	return &deps{
		productRepo: mocks.NewMockproductRepo(t),
		imageRepo:   mocks.NewMockimageRepo(t),
		now: func() time.Time {
			return time.Time{}
		},
		idGenerator: func() string {
			return productID
		},
	}
}

func TestUsecase_GetProducts(t *testing.T) {
	ctx := context.Background()
	in := usecase.GetProductsIn{
		Page:  1,
		Limit: 10,
	}
	storageIn := storage.GetProductsIn{
		Limit:  10,
		Offset: 0,
	}
	productIDs := []string{
		productID,
		"01900000-0000-7000-8000-000000000065",
	}

	tests := []struct {
		name    string
		want    []models.Product
		wantErr string
		setup   func(d *deps)
	}{
		{
			name: "success",
			want: []models.Product{
				{
					ID:          productID,
					Title:       "title",
					Description: "description",
					Price:       1000,
					Images:      nil,
				},
				{
					ID:          "01900000-0000-7000-8000-000000000065",
					Title:       "second",
					Description: "second description",
					Price:       2000,
					Images:      nil,
				},
			},
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					Get(mock.Anything, storageIn).
					Return([]models.Product{
						{
							ID:          productID,
							Title:       "title",
							Description: "description",
							Price:       1000,
						},
						{
							ID:          "01900000-0000-7000-8000-000000000065",
							Title:       "second",
							Description: "second description",
							Price:       2000,
						},
					}, nil).
					Once()
				d.imageRepo.EXPECT().
					GetByProductIDs(mock.Anything, productIDs).
					Return(nil, nil).
					Once()
			},
		},
		{
			name: "db error",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					Get(mock.Anything, storageIn).
					Return(nil, errors.New("some error")).
					Once()
			},
			wantErr: "get products: some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupDeps(t)
			tt.setup(deps)

			uc := product.New(deps.productRepo, deps.imageRepo, deps.now, deps.idGenerator)

			products, err := uc.GetProducts(ctx, in)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, products)
			}
		})
	}
}

func TestUsecase_GetProductByID(t *testing.T) {
	ctx := context.Background()
	p := models.Product{
		ID:          productID,
		Title:       "title",
		Description: "description",
		Price:       1000,
	}

	tests := []struct {
		name    string
		wantErr string
		setup   func(d *deps)
	}{
		{
			name: "success",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					GetByID(mock.Anything, productID).
					Return(p, nil).
					Once()
				d.imageRepo.EXPECT().
					GetByProductID(mock.Anything, productID).
					Return(nil, nil).
					Once()
			},
		},
		{
			name: "not found",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					GetByID(mock.Anything, productID).
					Return(models.Product{}, sqlp.ErrNotFound).
					Once()
			},
			wantErr: usecase.ErrProductNotFound.Error(),
		},
		{
			name: "db error",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					GetByID(mock.Anything, productID).
					Return(models.Product{}, errors.New("some error")).
					Once()
			},
			wantErr: "get product by id: some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupDeps(t)
			tt.setup(deps)

			uc := product.New(deps.productRepo, deps.imageRepo, deps.now, deps.idGenerator)

			got, err := uc.GetProductByID(ctx, productID)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, p, got)
			}
		})
	}
}

func TestUsecase_CreateProduct(t *testing.T) {
	ctx := context.Background()
	in := usecase.CreateProductIn{
		Title:       "title",
		Description: "description",
		Price:       1000,
	}
	p := models.Product{
		ID:          productID,
		Title:       "title",
		Description: "description",
		Price:       1000,
	}

	tests := []struct {
		name    string
		wantErr string
		setup   func(d *deps)
	}{
		{
			name: "success",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					Create(mock.Anything, p).
					Return(nil).
					Once()
			},
		},
		{
			name: "db error",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					Create(mock.Anything, p).
					Return(errors.New("some error")).
					Once()
			},
			wantErr: "create product: some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupDeps(t)
			tt.setup(deps)

			uc := product.New(deps.productRepo, deps.imageRepo, deps.now, deps.idGenerator)

			id, err := uc.CreateProduct(ctx, in)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, productID, id)
			}
		})
	}
}

func TestUsecase_UpdateProduct(t *testing.T) {
	ctx := context.Background()
	in := usecase.UpdateProductIn{
		Title:       new("new title"),
		Description: new("new description"),
	}
	storageIn := storage.UpdateProductIn{
		Title:       in.Title,
		Description: in.Description,
	}

	tests := []struct {
		name    string
		wantErr string
		setup   func(d *deps)
	}{
		{
			name: "success",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					Update(mock.Anything, productID, storageIn).
					Return(nil).
					Once()
			},
		},
		{
			name: "not found",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					Update(mock.Anything, productID, storageIn).
					Return(sqlp.ErrNotFound).
					Once()
			},
			wantErr: usecase.ErrProductNotFound.Error(),
		},
		{
			name: "db error",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					Update(mock.Anything, productID, storageIn).
					Return(errors.New("some error")).
					Once()
			},
			wantErr: "update product: some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupDeps(t)
			tt.setup(deps)

			uc := product.New(deps.productRepo, deps.imageRepo, deps.now, deps.idGenerator)

			err := uc.UpdateProduct(ctx, productID, in)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUsecase_DeleteProduct(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		wantErr string
		setup   func(d *deps)
	}{
		{
			name: "success",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					Delete(mock.Anything, productID).
					Return(nil).
					Once()
			},
		},
		{
			name: "not found",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					Delete(mock.Anything, productID).
					Return(sqlp.ErrNotFound).
					Once()
			},
			wantErr: usecase.ErrProductNotFound.Error(),
		},
		{
			name: "db error",
			setup: func(d *deps) {
				d.productRepo.EXPECT().
					Delete(mock.Anything, productID).
					Return(errors.New("some error")).
					Once()
			},
			wantErr: "delete product: some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupDeps(t)
			tt.setup(deps)

			uc := product.New(deps.productRepo, deps.imageRepo, deps.now, deps.idGenerator)

			err := uc.DeleteProduct(ctx, productID)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
