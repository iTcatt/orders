package models

import "time"

// Product – модель товара
type Product struct {
	ID          int32     `db:"id"`          // ID – идентификатор товара
	Title       string    `db:"title"`       // Title – название товара
	Description string    `db:"description"` // Description – описание товара
	Price       int32     `db:"price"`       // Price – цена в копейках
	CreatedAt   time.Time `db:"created_at"`  // CreatedAt – дата создания товара
	UpdatedAt   time.Time `db:"updated_at"`  // UpdatedAt – дата обновления товара
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
