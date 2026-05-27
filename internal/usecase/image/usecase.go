package image

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"

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

func (u *Usecase) Reorder(ctx context.Context, productID string, positions []usecase.ImagePosition) error {
	images, err := u.imageRepo.GetByProductID(ctx, productID)
	if err != nil {
		return fmt.Errorf("get images: %w", err)
	}

	if len(positions) != len(images) {
		return fmt.Errorf("expected %d positions, got %d", len(images), len(positions))
	}

	imageMap := lo.KeyBy(images, func(img models.Image) string { return img.ID })
	updated := make([]models.Image, len(images))

	for _, p := range positions {
		img, ok := imageMap[p.ID]
		if !ok {
			return fmt.Errorf("image %s does not belong to product", p.ID)
		}
		if p.Position < 0 || int(p.Position) >= len(images) {
			return fmt.Errorf("position %d out of range [0, %d]", p.Position, len(images)-1)
		}
		// two different IDs can claim the same slot; zero value means the slot is free
		if updated[p.Position] != (models.Image{}) {
			return fmt.Errorf("duplicate position %d", p.Position)
		}
		img.Position = p.Position
		updated[p.Position] = img
	}

	return u.txManager.RunInTx(ctx, func(ctx context.Context) error {
		for _, img := range updated {
			if err := u.imageRepo.UpdatePosition(ctx, img); err != nil {
				return fmt.Errorf("update position for image %s: %w", img.ID, err)
			}
		}
		return nil
	})
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
