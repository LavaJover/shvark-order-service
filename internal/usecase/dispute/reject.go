package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
)

func (disputeUc *DefaultDisputeUsecase) RejectDispute(disputeID string) error {
	dispute, err := disputeUc.disputeRepo.GetDisputeByID(disputeID)
	if err != nil {
		return err
	}
	if dispute.Status != domain.DisputeOpened && dispute.Status != domain.DisputeFreezed {
		return fmt.Errorf("invalid dispute status to reject dispute: %s", dispute.Status)
	}
	order, err := disputeUc.orderRepo.GetOrderByID(dispute.OrderID)
	if err != nil {
		return err
	}
	if order.Status != domain.StatusDisputeCreated {
		return fmt.Errorf("invalid order status to reject dispute: %s", order.Status)
	}
	op := &DisputeOperation{
		OrderID: order.ID,
		DisputeID: disputeID,
		Operation: "reject",
		OldDisputeStatus: dispute.Status,
		NewDisputeStatus: domain.DisputeRejected,
		OldOrderStatus: order.Status,
		NewOrderStatus: dispute.OrderStatusOriginal,
		WalletOp: &WalletOperation{
			Type: "release",
			Request: walletRequest.ReleaseRequest{
				TraderID: order.RequisiteDetails.TraderID,
				MerchantID: order.MerchantInfo.MerchantID,
				OrderID: fmt.Sprintf("%s_dispute_%s", dispute.OrderID, dispute.ID),
				RewardPercent: 1,
				PlatformFee: 1,
			},
		},
		CreatedAt: time.Now(),
	}
	if err := disputeUc.ProcessDisputeOperation(context.Background(), op); err != nil {
		return err
	}
	if order.CallbackUrl != "" {
		notifier.SendCallback(
			order.CallbackUrl,
			order.MerchantInfo.MerchantOrderID,
			string(domain.StatusCanceled),
			0, 0, 0,
		)
	}
	
	return nil
}