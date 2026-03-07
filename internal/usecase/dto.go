package usecase

type CreateProductIn struct {
	Title       string
	Description string
	Price       int32
}

type UpdateProductIn struct {
	Title       *string
	Description *string
	Price       *int32
}
