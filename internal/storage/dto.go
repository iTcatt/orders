package storage

type UpdateProductIn struct {
	Title       *string
	Description *string
	Price       *int32
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
	return result
}
