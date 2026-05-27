package category

type handler struct {
	uc categoryUsecase
}

func New(uc categoryUsecase) *handler {
	return &handler{uc: uc}
}
