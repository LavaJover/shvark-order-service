package usecase

import (
	"context"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
)

func (uc *DefaultOrderUsecase) AcceptOrder(orderID string) error {
	// Find exact order
	order, err := uc.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != domain.StatusCreated {
		return domain.ErrResolveDisputeFailed
	}

	op := &OrderOperation{
		OrderID:   orderID,
		Operation: "accept",
		OldStatus: domain.StatusCreated,
		NewStatus: domain.StatusPending,
		WalletOp: nil,
		CreatedAt: time.Now(),
	}

	if err := uc.ProcessOrderOperation(context.Background(), op); err != nil {
		return err
	}

	return nil
}