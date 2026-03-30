package usecase

type GetProductsIn struct {
	Page  uint32
	Limit uint32
}

type CreateProductIn struct {
	Title       string
	Description string
	Price       uint32
}

type UpdateProductIn struct {
	Title       *string
	Description *string
	Price       *uint32
}
