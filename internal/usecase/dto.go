package usecase

import "io"

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

type UploadImageIn struct {
	ProductID   string
	File        io.ReadCloser
	Size        int64  // file size in bytes
	ContentType string // MIME type, e.g. "image/jpeg"
	Ext         string // file extension without dot, e.g. "jpg" — used to build the object store key
}
