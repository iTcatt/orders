package usecase

import "io"

type GetProductsIn struct {
	Page       uint32
	Limit      uint32
	CategoryID *int
}

type CreateProductIn struct {
	Title       string
	Description string
	Price       uint32
	CategoryID  int
}

type UpdateProductIn struct {
	Title       *string
	Description *string
	Price       *uint32
	CategoryID  *int
}

type ImagePosition struct {
	ID       string
	Position int16
}

type UploadImageIn struct {
	ProductID   string
	File        io.ReadCloser
	Size        int64  // file size in bytes
	ContentType string // MIME type, e.g. "image/jpeg"
	Ext         string // file extension without dot, e.g. "jpg" — used to build the object store key
}
