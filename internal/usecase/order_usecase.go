package usecase

import "github.com/LavaJover/shvark-order-service/internal/domain"

type DefaultOrderUsecase struct {
	OrderRepo 		domain.OrderRepository
	BankDetailRepo 	domain.BankDetailRepository
}

func NewDefaultOrderUsecase(orderRepo domain.OrderRepository, bankDetailRepo domain.BankDetailRepository) *DefaultOrderUsecase {
	return &DefaultOrderUsecase{
		OrderRepo: orderRepo,
		BankDetailRepo: bankDetailRepo,
	}
}

func (uc *DefaultOrderUsecase) CreateOrder(order *domain.Order, bankDetail *domain.BankDetail) (string, error) {
	if err := uc.BankDetailRepo.SaveBankDetail(bankDetail); err != nil {
		return "", err
	}
	return uc.OrderRepo.CreateOrder(order)
}

func (uc *DefaultOrderUsecase) ApproveOrder(orderID string) error {
	return uc.OrderRepo.UpdateOrderStatus(orderID, "COMPLETED")
}

func (uc *DefaultOrderUsecase) CancelOrder(orderID string) error {
	return uc.OrderRepo.UpdateOrderStatus(orderID, "FAILED")
}

func (uc *DefaultOrderUsecase) GetOrderByID(orderID string) (*domain.Order, error) {
	return uc.OrderRepo.GetOrderByID(orderID)
}

func (uc *DefaultOrderUsecase) GetOrdersByTraderID(traderID string) ([]*domain.Order, error) {
	return uc.OrderRepo.GetOrdersByTraderID(traderID)
}