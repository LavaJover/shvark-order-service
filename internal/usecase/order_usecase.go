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

func (uc *DefaultOrderUsecase) ApproveOrder(orderID string) error {
	return uc.Repo.UpdateOrderStatus(orderID, "COMPLETED")
}

func (uc *DefaultOrderUsecase) CancelOrder(orderID string) error {
	return uc.Repo.UpdateOrderStatus(orderID, "FAILED")
}

func (uc *DefaultOrderUsecase) GetOrderByID(orderID string) (*domain.Order, error) {
	return uc.Repo.GetOrderByID(orderID)
}

func (uc *DefaultOrderUsecase) GetOrdersByTraderID(traderID string) ([]*domain.Order, error) {
	return uc.Repo.GetOrdersByTraderID(traderID)
}