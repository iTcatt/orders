package dto

type CreateProductIn struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int32  `json:"price"`
}

type UpdateProductIn struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Price       *int32  `json:"price,omitempty"`
}
