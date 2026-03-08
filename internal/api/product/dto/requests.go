package dto

type CreateProductIn struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"required"`
	Price       int32  `json:"price" validate:"required,gt=0"`
}

type UpdateProductIn struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty"`
	Price       *int32  `json:"price,omitempty" validate:"omitempty,gt=0"`
}
