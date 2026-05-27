package storage

type UpdateProductIn struct {
	Title       *string
	Description *string
	Price       *uint32
	CategoryID  *int
}

func (in UpdateProductIn) ToMap() map[string]any {
	result := make(map[string]any)
	if in.Title != nil {
		result["title"] = *in.Title
	}
	if in.Description != nil {
		result["description"] = *in.Description
	}
	if in.Price != nil {
		result["price"] = *in.Price
	}
	if in.CategoryID != nil {
		result["category_id"] = *in.CategoryID
	}
	return result
}

type GetProductsIn struct {
	Limit      uint32
	Offset     uint32
	CategoryID *int
}
