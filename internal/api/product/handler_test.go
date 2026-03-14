package product_test

import (
	"testing"

	"iTcatt/orders/internal/api/product/mocks"
)

type deps struct {
	uc *mocks.MockproductUsecase
}

func setupDeps(t *testing.T) deps {
	return deps{
		uc: mocks.NewMockproductUsecase(t),
	}
}
