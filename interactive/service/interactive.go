package service

type InteractiveService interface {
}

type interactiveService struct {
}

func NewInteractiveService() InteractiveService {
	return &interactiveService{}
}
