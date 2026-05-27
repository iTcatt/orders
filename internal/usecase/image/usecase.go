package image

import (
	"context"
	"errors"
	"fmt"
	"time"

	"iTcatt/orders/internal/models"
	"iTcatt/orders/internal/usecase"
	"iTcatt/orders/pkg/sqlp"
)

const maxImagesPerProduct = 3

var ErrTooManyImages = errors.New("product already has maximum number of images")

type Usecase struct {
	imageRepo   imageRepo
	objectStore objectStore
	productRepo productRepo
	txManager   txManager
	now         func() time.Time
	idGenerator func() string
}

func New(
	imageRepo imageRepo,
	objectStore objectStore,
	productRepo productRepo,
	txManager txManager,
	now func() time.Time,
	idGen func() string,
) *Usecase {
	return &Usecase{
		imageRepo:   imageRepo,
		objectStore: objectStore,
		productRepo: productRepo,
		txManager:   txManager,
		now:         now,
		idGenerator: idGen,
	}
}

// Upload uploads an image file for a product and saves metadata to the database.
func (u *Usecase) Upload(ctx context.Context, input usecase.UploadImageIn) (models.Image, error) {
	if _, err := u.productRepo.GetByID(ctx, input.ProductID); err != nil {
		if errors.Is(err, sqlp.ErrNotFound) {
			return models.Image{}, usecase.ErrProductNotFound
		}
		return models.Image{}, fmt.Errorf("get product: %w", err)
	}

	count, err := u.imageRepo.Count(ctx, input.ProductID)
	if err != nil {
		return models.Image{}, fmt.Errorf("count images: %w", err)
	}
	if count >= maxImagesPerProduct {
		return models.Image{}, ErrTooManyImages
	}

	imageID := u.idGenerator()
	key := fmt.Sprintf("products/%s/%s.%s", input.ProductID, imageID, input.Ext)

	img := models.Image{
		ID:        imageID,
		ProductID: input.ProductID,
		URL:       u.objectStore.URL(key),
		ObjectKey: key,
		Position:  int16(count), //nolint:gosec // count < maxImagesPerProduct (3), no overflow
		CreatedAt: u.now(),
	}

	if err := u.txManager.RunInTx(ctx, func(ctx context.Context) error {
		if err := u.imageRepo.Create(ctx, img); err != nil {
			return fmt.Errorf("save image: %w", err)
		}

		if err := u.objectStore.Upload(ctx, key, input.File, input.Size, input.ContentType); err != nil {
			return fmt.Errorf("upload to object store: %w", err)
		}

		return nil
	}); err != nil {
		return models.Image{}, err
	}

	return img, nil
}

func (u *Usecase) Delete(ctx context.Context, imageID string) error {
	img, err := u.imageRepo.Get(ctx, imageID)
	if err != nil {
		if errors.Is(err, sqlp.ErrNotFound) {
			return usecase.ErrImageNotFound
		}
		return fmt.Errorf("get image: %w", err)
	}

	return u.txManager.RunInTx(ctx, func(ctx context.Context) error {
		if err := u.imageRepo.Delete(ctx, imageID); err != nil {
			return fmt.Errorf("delete image record: %w", err)
		}

		if err := u.objectStore.Delete(ctx, img.ObjectKey); err != nil {
			return fmt.Errorf("delete from object store: %w", err)
		}

		return nil
	})
}
