package usecase

import (
	"context"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (disputeUc *DefaultDisputeUsecase) FreezeDispute(disputeID string) error {
	dispute, err := disputeUc.disputeRepo.GetDisputeByID(disputeID)
	if err != nil {
		return err
	}
	if dispute.Status != domain.DisputeOpened {
		return status.Error(codes.FailedPrecondition, "dispute is not opened yet")
	}

	op := &DisputeOperation{
		OrderID: dispute.OrderID,
		OldOrderStatus: domain.StatusDisputeCreated,
		NewOrderStatus: domain.StatusDisputeCreated,
		DisputeID: dispute.ID,
		Operation: "freeze",
		OldDisputeStatus: dispute.Status,
		NewDisputeStatus: domain.DisputeFreezed,
		WalletOp: nil,
		CreatedAt: time.Now(),
	}

	return disputeUc.ProcessDisputeOperation(context.Background(), op)
}
