package product

type handler struct {
	uc productUsecase
}

func New(uc productUsecase) *handler {
	return &handler{uc: uc}
}
