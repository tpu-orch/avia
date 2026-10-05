package service

type AviaRepository interface{}

type AviaService struct {
	repo AviaRepository
}

func NewAviaService(repo AviaRepository) *AviaService {
	return &AviaService{repo: repo}
}
