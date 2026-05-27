package image_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/internal/usecase/image"
	"iTcatt/orders/internal/usecase/image/mocks"
	"iTcatt/orders/pkg/sqlp"
)

const (
	productID = "01900000-0000-7000-8000-000000000064"
	imageID   = "01900000-0000-7000-8000-000000000065"
)

type deps struct {
	imageRepo   *mocks.MockimageRepo
	objectStore *mocks.MockobjectStore
	productRepo *mocks.MockproductRepo
	txManager   *mocks.MocktxManager
	now         func() time.Time
	idGenerator func() string
}

func setupDeps(t *testing.T) *deps {
	t.Helper()
	return &deps{
		imageRepo:   mocks.NewMockimageRepo(t),
		objectStore: mocks.NewMockobjectStore(t),
		productRepo: mocks.NewMockproductRepo(t),
		txManager:   mocks.NewMocktxManager(t),
		now:         func() time.Time { return time.Time{} },
		idGenerator: func() string { return imageID },
	}
}

func (d *deps) newUsecase() *image.Usecase {
	return image.New(d.imageRepo, d.objectStore, d.productRepo, d.txManager, d.now, d.idGenerator)
}

func executeTx(d *deps) {
	d.txManager.EXPECT().
		RunInTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
}

func TestUsecase_Upload(t *testing.T) {
	ctx := context.Background()

	file := io.NopCloser(strings.NewReader("data"))
	in := usecase.UploadImageIn{
		ProductID:   productID,
		File:        file,
		Size:        4,
		ContentType: "image/jpeg",
		Ext:         "jpg",
	}

	objectKey := "products/" + productID + "/" + imageID + ".jpg"
	imageURL := "http://minio/products/" + productID + "/" + imageID + ".jpg"
	wantImg := models.Image{
		ID:        imageID,
		ProductID: productID,
		URL:       imageURL,
		ObjectKey: objectKey,
		Position:  0,
		CreatedAt: time.Time{},
	}

	tests := []struct {
		name    string
		wantErr string
		setup   func(d *deps)
	}{
		{
			name: "success",
			setup: func(d *deps) {
				d.productRepo.EXPECT().GetByID(mock.Anything, productID).Return(models.Product{}, nil).Once()
				d.imageRepo.EXPECT().Count(mock.Anything, productID).Return(0, nil).Once()
				d.objectStore.EXPECT().URL(objectKey).Return(imageURL).Once()
				executeTx(d)
				d.imageRepo.EXPECT().Create(mock.Anything, wantImg).Return(nil).Once()
				d.objectStore.EXPECT().Upload(mock.Anything, objectKey, file, int64(4), "image/jpeg").Return(nil).Once()
			},
		},
		{
			name: "product not found",
			setup: func(d *deps) {
				d.productRepo.EXPECT().GetByID(mock.Anything, productID).Return(models.Product{}, sqlp.ErrNotFound).Once()
			},
			wantErr: usecase.ErrProductNotFound.Error(),
		},
		{
			name: "product repo error",
			setup: func(d *deps) {
				d.productRepo.EXPECT().GetByID(mock.Anything, productID).Return(models.Product{}, errors.New("db error")).Once()
			},
			wantErr: "get product: db error",
		},
		{
			name: "count images error",
			setup: func(d *deps) {
				d.productRepo.EXPECT().GetByID(mock.Anything, productID).Return(models.Product{}, nil).Once()
				d.imageRepo.EXPECT().Count(mock.Anything, productID).Return(0, errors.New("db error")).Once()
			},
			wantErr: "count images: db error",
		},
		{
			name: "too many images",
			setup: func(d *deps) {
				d.productRepo.EXPECT().GetByID(mock.Anything, productID).Return(models.Product{}, nil).Once()
				d.imageRepo.EXPECT().Count(mock.Anything, productID).Return(3, nil).Once()
			},
			wantErr: image.ErrTooManyImages.Error(),
		},
		{
			name: "image repo create error",
			setup: func(d *deps) {
				d.productRepo.EXPECT().GetByID(mock.Anything, productID).Return(models.Product{}, nil).Once()
				d.imageRepo.EXPECT().Count(mock.Anything, productID).Return(0, nil).Once()
				d.objectStore.EXPECT().URL(objectKey).Return(imageURL).Once()
				executeTx(d)
				d.imageRepo.EXPECT().Create(mock.Anything, wantImg).Return(errors.New("db error")).Once()
			},
			wantErr: "save image: db error",
		},
		{
			name: "object store upload error",
			setup: func(d *deps) {
				d.productRepo.EXPECT().GetByID(mock.Anything, productID).Return(models.Product{}, nil).Once()
				d.imageRepo.EXPECT().Count(mock.Anything, productID).Return(0, nil).Once()
				d.objectStore.EXPECT().URL(objectKey).Return(imageURL).Once()
				executeTx(d)
				d.imageRepo.EXPECT().Create(mock.Anything, wantImg).Return(nil).Once()
				d.objectStore.EXPECT().Upload(mock.Anything, objectKey, file, int64(4), "image/jpeg").Return(errors.New("minio error")).Once()
			},
			wantErr: "upload to object store: minio error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := setupDeps(t)
			tt.setup(d)

			got, err := d.newUsecase().Upload(ctx, in)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, wantImg, got)
			}
		})
	}
}

func TestUsecase_Delete(t *testing.T) {
	ctx := context.Background()

	objectKey := "products/" + productID + "/" + imageID + ".jpg"
	img := models.Image{
		ID:        imageID,
		ProductID: productID,
		ObjectKey: objectKey,
	}

	tests := []struct {
		name    string
		wantErr string
		setup   func(d *deps)
	}{
		{
			name: "success",
			setup: func(d *deps) {
				d.imageRepo.EXPECT().Get(mock.Anything, imageID).Return(img, nil).Once()
				executeTx(d)
				d.imageRepo.EXPECT().Delete(mock.Anything, imageID).Return(nil).Once()
				d.objectStore.EXPECT().Delete(mock.Anything, objectKey).Return(nil).Once()
			},
		},
		{
			name: "get image error",
			setup: func(d *deps) {
				d.imageRepo.EXPECT().Get(mock.Anything, imageID).Return(models.Image{}, errors.New("db error")).Once()
			},
			wantErr: "get image: db error",
		},
		{
			name: "image not found",
			setup: func(d *deps) {
				d.imageRepo.EXPECT().Get(mock.Anything, imageID).Return(models.Image{}, sqlp.ErrNotFound).Once()
			},
			wantErr: usecase.ErrImageNotFound.Error(),
		},
		{
			name: "image repo delete error",
			setup: func(d *deps) {
				d.imageRepo.EXPECT().Get(mock.Anything, imageID).Return(img, nil).Once()
				executeTx(d)
				d.imageRepo.EXPECT().Delete(mock.Anything, imageID).Return(errors.New("db error")).Once()
			},
			wantErr: "delete image record: db error",
		},
		{
			name: "object store delete error",
			setup: func(d *deps) {
				d.imageRepo.EXPECT().Get(mock.Anything, imageID).Return(img, nil).Once()
				executeTx(d)
				d.imageRepo.EXPECT().Delete(mock.Anything, imageID).Return(nil).Once()
				d.objectStore.EXPECT().Delete(mock.Anything, objectKey).Return(errors.New("minio error")).Once()
			},
			wantErr: "delete from object store: minio error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := setupDeps(t)
			tt.setup(d)

			err := d.newUsecase().Delete(ctx, imageID)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
