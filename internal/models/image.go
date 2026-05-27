package models

import "time"

type Image struct {
	ID        string    `db:"id"`
	ProductID string    `db:"product_id"`
	URL       string    `db:"url"`        // public URL to access the image
	ObjectKey string    `db:"object_key"` // path to the object in the object store, e.g. "products/{productID}/{imageID}.jpg"
	Position  int16     `db:"position"`   // display order among the product's images, zero-based
	CreatedAt time.Time `db:"created_at"`
}
