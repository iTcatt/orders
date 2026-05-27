package models

import "time"

type Product struct {
	ID          string    `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Price       uint32    `db:"price"` // price in smallest currency unit (kopecks/cents)
	Images      []Image   `db:"-"`     // not stored in this table; populated separately
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (p *Product) ToMap() map[string]any {
	return map[string]any{
		"id":          p.ID,
		"title":       p.Title,
		"description": p.Description,
		"price":       p.Price,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
	}
}
