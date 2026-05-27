package usecase

import "errors"

var (
	ErrProductNotFound = errors.New("product not found")
	ErrImageNotFound   = errors.New("image not found")
)
