package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
)

func (disputeUc *DefaultDisputeUsecase) AcceptDispute(disputeID string) error {
	dispute, err := disputeUc.disputeRepo.GetDisputeByID(disputeID)
	if err != nil {
		return err
	}
	order, err := disputeUc.orderRepo.GetOrderByID(dispute.OrderID)
	if err != nil {
		return err
	}
	if order.Status != domain.StatusDisputeCreated {
		return fmt.Errorf("invalid order status to accept dispute: %s", order.Status)
	}
	traffic, err := disputeUc.trafficRepo.GetTrafficByTraderMerchant(order.RequisiteDetails.TraderID, order.MerchantInfo.MerchantID)
	if err != nil {
		return err
	}
	// Search for team relations to find commission users
	var commissionUsers []walletRequest.CommissionUser
	teamRelations, err := disputeUc.teamRelationsUsecase.GetRelationshipsByTraderID(order.RequisiteDetails.TraderID)
	if err == nil {
		for _, teamRelation := range teamRelations {
			commissionUsers = append(commissionUsers, walletRequest.CommissionUser{
				UserID: teamRelation.TeamLeadID,
				Commission: teamRelation.TeamRelationshipRapams.Commission,
			})
		}
	}
	op := &DisputeOperation{
		OrderID: order.ID,
		DisputeID: disputeID,
		Operation: "accept",
		OldDisputeStatus: dispute.Status,
		NewDisputeStatus: domain.DisputeAccepted,
		OldOrderStatus: order.Status,
		NewOrderStatus: domain.StatusCompleted,
		NewOrderAmountFiat: dispute.DisputeAmountFiat,
		NewOrderAmountCrypto: dispute.DisputeAmountCrypto,
		NewOrderAmountCryptoRate: dispute.DisputeCryptoRate,		
		WalletOp: &WalletOperation{
			Type: "release",
			Request: walletRequest.ReleaseRequest{
				TraderID: order.RequisiteDetails.TraderID,
				MerchantID: order.MerchantInfo.MerchantID,
				OrderID: fmt.Sprintf("%s_dispute_%s", dispute.OrderID, dispute.ID),
				RewardPercent: traffic.TraderRewardPercent,
				PlatformFee: traffic.PlatformFee,
				CommissionUsers: commissionUsers,
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
			string(domain.StatusCompleted),
			dispute.DisputeAmountCrypto, dispute.DisputeAmountFiat, dispute.DisputeCryptoRate,
		)
	}
	return nil
}

func (disputeUc *DefaultDisputeUsecase) AcceptExpiredDisputes() error {
	disputes, err := disputeUc.disputeRepo.FindExpiredDisputes()
	if err != nil {
		return err
	}
	for _, dispute := range disputes {
		if err := disputeUc.AcceptDispute(dispute.ID); err != nil {
			log.Printf("failed to accept dispute %s\n", dispute.ID)
			return status.Error(codes.Internal, err.Error())
		}
	} 
	return nil
}