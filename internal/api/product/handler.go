package product

type Handler struct {
	storage productStorage
}

func New(s productStorage) *Handler {
	return &Handler{storage: s}
}
