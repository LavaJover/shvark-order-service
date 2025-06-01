package usecase

import "github.com/LavaJover/shvark-order-service/internal/domain"

type DefaultOrderUsecase struct {
	Repo domain.OrderRepository
}

func NewDefaultOrderUsecase(repo domain.OrderRepository) *DefaultOrderUsecase {
	return &DefaultOrderUsecase{Repo: repo}
}

func (uc *DefaultOrderUsecase) CreateOrder(order *domain.Order) (string, error) {
	return uc.Repo.CreateOrder(order)
}