package dto

import "time"

type Image struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type Product struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       uint32    `json:"price"`
	CategoryID  int       `json:"category_id"`
	Images      []Image   `json:"images"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProductOut struct {
	ID string `json:"id"`
}
