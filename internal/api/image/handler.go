package image

type handler struct {
	uc imageUsecase
}

func New(uc imageUsecase) *handler {
	return &handler{uc: uc}
}
