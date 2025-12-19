package usecase

import (
	"context"
	"log/slog"
	"time"

	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	publisher "github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (uc *DefaultOrderUsecase) ApproveOrder(orderID string) error {
	// Find exact order
	order, err := uc.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != domain.StatusPending {
		return domain.ErrResolveDisputeFailed
	}

	if order.Type == domain.TypePayIn {
		return uc.processPayInApprove(order)
	} else if order.Type == domain.TypePayOut {
		return uc.processPayOutApprove(order)
	}

	return status.Errorf(codes.Internal, "failed to approve order: unknown order type")
}

func (uc *DefaultOrderUsecase) processPayInApprove(order *domain.Order) error {
	orderID := order.ID
	// Search for team relations to find commission users
	var commissionUsers []walletRequest.CommissionUser
	teamRelations, err := uc.TeamRelationsUsecase.GetRelationshipsByTraderID(order.RequisiteDetails.TraderID)
	if err == nil {
		for _, teamRelation := range teamRelations {
			commissionUsers = append(commissionUsers, walletRequest.CommissionUser{
				UserID: teamRelation.TeamLeadID,
				Commission: teamRelation.TeamRelationshipRapams.Commission,
			})
		}
	}
	op := &OrderOperation{
		OrderID:   orderID,
		Operation: "approve",
		OldStatus: domain.StatusPending,
		NewStatus: domain.StatusCompleted,
		WalletOp: &WalletOperation{
			Type: "release",
			Request: walletRequest.ReleaseRequest{
				TraderID:        order.RequisiteDetails.TraderID,
				MerchantID:      order.MerchantInfo.MerchantID,
				OrderID:         order.ID,
				RewardPercent:   order.TraderReward,
				PlatformFee:     order.PlatformFee,
				CommissionUsers: commissionUsers,
			},
		},
		CreatedAt: time.Now(),
	}

	if err := uc.ProcessOrderOperation(context.Background(), op); err != nil {
		return err
	}

	go func(event publisher.OrderEvent){
		if err := uc.Publisher.PublishOrder(event); err != nil {
			slog.Error("failed to publish kafka OrderEvent", "stage", "approving", "error", err.Error())
		}
	}(publisher.OrderEvent{
		OrderID: order.ID,
		TraderID: order.RequisiteDetails.TraderID,
		Status: "✅Сделка закрыта",
		AmountFiat: order.AmountInfo.AmountFiat,
		Currency: order.AmountInfo.Currency,
		BankName: order.RequisiteDetails.BankName,
		Phone: order.RequisiteDetails.Phone,
		CardNumber: order.RequisiteDetails.CardNumber,
		Owner: order.RequisiteDetails.Owner,
	})

	if order.CallbackUrl != "" {
		notifier.SendCallback(
			order.CallbackUrl,
			order.MerchantInfo.MerchantOrderID,
			string(domain.StatusCompleted),
			0, 0, 0,
		)
	}

	// ✅ ЗАПИСЬ МЕТРИКИ ЗАВЕРШЕННОГО ЗАКАЗА (с payment_system)
	uc.recordOrderCompletedMetrics(order, order.RequisiteDetails.PaymentSystem)

	return nil
}

func (uc *DefaultOrderUsecase) processPayOutApprove(order *domain.Order) error {
	orderID := order.ID
	// Search for team relations to find commission users
	var commissionUsers []walletRequest.CommissionUser
	// teamRelations, err := uc.TeamRelationsUsecase.GetRelationshipsByTraderID(order.RequisiteDetails.TraderID)
	// if err == nil {
	// 	for _, teamRelation := range teamRelations {
	// 		commissionUsers = append(commissionUsers, walletRequest.CommissionUser{
	// 			UserID: teamRelation.TeamLeadID,
	// 			Commission: teamRelation.TeamRelationshipRapams.Commission,
	// 		})
	// 	}
	// }
	op := &OrderOperation{
		OrderID:   orderID,
		Operation: "approve",
		OldStatus: domain.StatusPending,
		NewStatus: domain.StatusCompleted,
		WalletOp: &WalletOperation{
			Type: "release",
			Request: walletRequest.ReleaseRequest{
				TraderID:        order.RequisiteDetails.TraderID,
				MerchantID:      order.MerchantInfo.MerchantID,
				OrderID:         order.ID,
				RewardPercent:   (1 + order.TraderReward),
				PlatformFee:     (1 + order.PlatformFee),
				CommissionUsers: commissionUsers,
			},
		},
		CreatedAt: time.Now(),
	}

	if err := uc.ProcessOrderOperation(context.Background(), op); err != nil {
		return err
	}

	go func(event publisher.OrderEvent){
		if err := uc.Publisher.PublishOrder(event); err != nil {
			slog.Error("failed to publish kafka OrderEvent", "stage", "approving", "error", err.Error())
		}
	}(publisher.OrderEvent{
		OrderID: order.ID,
		TraderID: order.RequisiteDetails.TraderID,
		Status: "✅Выплата завершена",
		AmountFiat: order.AmountInfo.AmountFiat,
		Currency: order.AmountInfo.Currency,
		BankName: order.RequisiteDetails.BankName,
		Phone: order.RequisiteDetails.Phone,
		CardNumber: order.RequisiteDetails.CardNumber,
		Owner: order.RequisiteDetails.Owner,
	})

	if order.CallbackUrl != "" {
		notifier.SendCallback(
			order.CallbackUrl,
			order.MerchantInfo.MerchantOrderID,
			string(domain.StatusCompleted),
			0, 0, 0,
		)
	}

	// ✅ ЗАПИСЬ МЕТРИКИ ЗАВЕРШЕННОГО ЗАКАЗА (с payment_system)
	uc.recordOrderCompletedMetrics(order, order.RequisiteDetails.PaymentSystem)

	return nil
}