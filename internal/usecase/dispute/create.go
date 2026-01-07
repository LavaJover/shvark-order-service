package usecase

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	publisher "github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	disputedto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/dispute"
	"github.com/jaevor/go-nanoid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Диспут открыт -> запись в БД со статусом DISPUTE_OPENED
// AutoAcceptAt -> в данное время система автоматически одобрит диспут
func (disputeUc *DefaultDisputeUsecase) CreateDispute(input *disputedto.CreateDisputeInput) error {
	order, err := disputeUc.orderRepo.GetOrderByID(input.OrderID)
	if err != nil {
		return err
	}
	if (order.Status != domain.StatusCanceled) && (order.Status != domain.StatusCompleted) {
		return status.Error(codes.FailedPrecondition, "invalid order status")
	}
	idGenerator, err := nanoid.Standard(15)
	if err != nil {
		return err
	}
	if order.MerchantInfo.MerchantID == "b5a09b96-8a99-48a6-a9e5-fe6033d374bd" {
        input.DisputeCryptoRate *= 1.1
        input.DisputeAmountCrypto /= 1.1
    }
    
	dispute := domain.Dispute{
		ID: idGenerator(),
		OrderID: input.OrderID,
		OrderStatusOriginal: order.Status,
		DisputeAmountFiat: input.DisputeAmountFiat,
		DisputeAmountCrypto: input.DisputeAmountCrypto,
		DisputeCryptoRate: input.DisputeCryptoRate,
		ProofUrl: input.ProofUrl,
		Reason: input.Reason,
		Status: domain.DisputeOpened,
		Ttl: input.Ttl,
		AutoAcceptAt: time.Now().Add(input.Ttl),
	}

	err = disputeUc.disputeRepo.CreateDispute(&dispute)
	if err != nil {
		return err
	}
	go func(event publisher.DisputeEvent){
		if err := disputeUc.kafkaPublisher.PublishDispute(event); err != nil {
			slog.Error("failed to publish kafka dispute event", "stage", "creating", "error",err.Error())
		}
	}(publisher.DisputeEvent{
		DisputeID: dispute.ID,
		OrderID: dispute.OrderID,
		TraderID: order.RequisiteDetails.TraderID,
		OrderAmountFiat: order.AmountInfo.AmountFiat,
		DisputeAmountFiat: dispute.DisputeAmountFiat,
		ProofUrl: dispute.ProofUrl,
		Reason: dispute.Reason,
		Status: "🆘Открыт диспут",
		BankName: order.RequisiteDetails.BankName,
		Phone: order.RequisiteDetails.Phone,
		CardNumber: order.RequisiteDetails.CardNumber,
		Owner: order.RequisiteDetails.Owner,
	})

	if order.Status == domain.StatusCompleted {
		err = disputeUc.walletHandler.Freeze(order.RequisiteDetails.TraderID, fmt.Sprintf("%s_dispute_%s", dispute.OrderID, dispute.ID), dispute.DisputeAmountCrypto-order.AmountInfo.AmountCrypto)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}else {
		err = disputeUc.walletHandler.Freeze(order.RequisiteDetails.TraderID, fmt.Sprintf("%s_dispute_%s", dispute.OrderID, dispute.ID), dispute.DisputeAmountCrypto)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
	err = disputeUc.orderRepo.UpdateOrderStatus(order.ID, domain.StatusDisputeCreated)
	if err != nil {
		return err
	}
	if order.CallbackUrl != "" {
		notifier.SendCallback(
			order.CallbackUrl,
			order.MerchantInfo.MerchantOrderID,
			string(domain.StatusDisputeCreated),
			0, 0, 0,
		)
	}
	return nil
}