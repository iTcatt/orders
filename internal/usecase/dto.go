package usecase

type GetProductsIn struct {
	Page  int32
	Limit int32
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
