package product

import "github.com/go-playground/validator/v10"

type handler struct {
	uc productUsecase
	v  *validator.Validate
}

func New(uc productUsecase) *handler {
	return &handler{
		uc: uc,
		v:  validator.New(),
	}
}
