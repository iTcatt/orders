package product

type Handler struct {
	uc productUsecase
}

func New(uc productUsecase) *Handler {
	return &Handler{uc: uc}
}
