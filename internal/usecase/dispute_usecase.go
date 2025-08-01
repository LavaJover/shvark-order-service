package usecase

import (
	"log"
	"time"

	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	disputedto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/dispute"
	"github.com/jaevor/go-nanoid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DisputeUsecase interface {
	CreateDispute(input *disputedto.CreateDisputeInput) error
	AcceptDispute(disputeID string) error
	RejectDispute(disputeID string) error
	FreezeDispute(disputeID string) error
	GetDisputeByID(disputeID string) (*domain.Dispute, error)
	GetDisputeByOrderID(orderID string) (*domain.Dispute, error)
	AcceptExpiredDisputes() error
	GetOrderDisputes(page, limit int64, status string) ([]*domain.Dispute, int64, error)
}

type DefaultDisputeUsecase struct {
	disputeRepo domain.DisputeRepository
	walletHandler *handlers.HTTPWalletHandler
	orderRepo domain.OrderRepository
	trafficRepo domain.TrafficRepository
	kafkaPublisher *kafka.KafkaPublisher
	teamRelationsUsecase TeamRelationsUsecase
	bankDetailRepo domain.BankDetailRepository
}

func NewDefaultDisputeUsecase(
	disputeRepo domain.DisputeRepository,
	walletHandler *handlers.HTTPWalletHandler,
	orderRepo domain.OrderRepository,
	trafficRepo domain.TrafficRepository,
	kafkaPublisher *kafka.KafkaPublisher,
	teamRelationsUsecase TeamRelationsUsecase,
	bankDetailRepo domain.BankDetailRepository,
	) *DefaultDisputeUsecase {
	return &DefaultDisputeUsecase{
		disputeRepo: disputeRepo,
		walletHandler: walletHandler,
		orderRepo: orderRepo,
		trafficRepo: trafficRepo,
		kafkaPublisher: kafkaPublisher,
		teamRelationsUsecase: teamRelationsUsecase,
		bankDetailRepo: bankDetailRepo,
	}
}

// Диспут открыт -> запись в БД со статусом DISPUTE_OPENED
// AutoAcceptAt -> в данное время система автоматически одобрит диспут
func (disputeUc *DefaultDisputeUsecase) CreateDispute(input *disputedto.CreateDisputeInput) error {
	order, err := disputeUc.orderRepo.GetOrderByID(input.OrderID)
	if err != nil {
		return err
	}
	// Если сделка не ушла в отмену, то диспут не открыть
	if order.Status != domain.StatusCanceled {
		return status.Error(codes.FailedPrecondition, "order is not even canceled")
	}
	idGenerator, err := nanoid.Standard(15)
	if err != nil {
		return err
	}
	dispute := domain.Dispute{
		ID: idGenerator(),
		OrderID: input.OrderID,
		DisputeAmountFiat: input.DisputeAmountFiat,
		DisputeAmountCrypto: input.DisputeAmountCrypto,
		DisputeCryptoRate: input.DisputeCryptoRate,
		ProofUrl: input.ProofUrl,
		Reason: input.Reason,
		Status: domain.DisputeOpened,
		Ttl: input.Ttl,
		AutoAcceptAt: time.Now().Add(input.Ttl),
	}

	bankDetailID := order.BankDetailID
	bankDetail, err := disputeUc.bankDetailRepo.GetBankDetailByID(bankDetailID)
	if err != nil {
		return err
	}
	err = disputeUc.disputeRepo.CreateDispute(&dispute)
	if err != nil {
		return err
	}
	disputeUc.kafkaPublisher.PublishDispute(kafka.DisputeEvent{
		DisputeID: dispute.ID,
		OrderID: dispute.OrderID,
		TraderID: bankDetail.TraderID,
		OrderAmountFiat: order.AmountInfo.AmountFiat,
		DisputeAmountFiat: dispute.DisputeAmountFiat,
		ProofUrl: dispute.ProofUrl,
		Reason: dispute.Reason,
		Status: "🆘Открыт диспут",
		BankName: bankDetail.BankName,
		Phone: bankDetail.Phone,
		CardNumber: bankDetail.CardNumber,
		Owner: bankDetail.Owner,
	})
	err = disputeUc.walletHandler.Freeze(bankDetail.TraderID, dispute.OrderID, dispute.DisputeAmountCrypto)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	err = disputeUc.orderRepo.UpdateOrderStatus(order.ID, domain.OrderStatus(domain.DisputeOpened))
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

func (disputeUc *DefaultDisputeUsecase) AcceptDispute(disputeID string) error {
	err :=  disputeUc.disputeRepo.UpdateDisputeStatus(disputeID, domain.DisputeAccepted)
	if err != nil {
		return err
	}
	dispute, err := disputeUc.disputeRepo.GetDisputeByID(disputeID)
	if err != nil {
		return err
	}
	order, err := disputeUc.orderRepo.GetOrderByID(dispute.OrderID)
	if err != nil {
		return err
	}
	bankDetailID := order.BankDetailID
	bankDetail, err := disputeUc.bankDetailRepo.GetBankDetailByID(bankDetailID)
	if err != nil {
		return err
	}
	traffic, err := disputeUc.trafficRepo.GetTrafficByTraderMerchant(bankDetail.TraderID, order.MerchantInfo.MerchantID)
	if err != nil {
		return err
	}
	// Search for team relations to find commission users
	var commissionUsers []walletRequest.CommissionUser
	teamRelations, err := disputeUc.teamRelationsUsecase.GetRelationshipsByTraderID(bankDetail.TraderID)
	if err == nil {
		for _, teamRelation := range teamRelations {
			commissionUsers = append(commissionUsers, walletRequest.CommissionUser{
				UserID: teamRelation.TeamLeadID,
				Commission: teamRelation.TeamRelationshipRapams.Commission,
			})
		}
	}
	releaseRequest := walletRequest.ReleaseRequest{
		TraderID: bankDetail.TraderID,
		MerchantID: order.MerchantInfo.MerchantID,
		OrderID: order.ID,
		RewardPercent: traffic.TraderRewardPercent,
		PlatformFee: traffic.PlatformFee,
		CommissionUsers: commissionUsers,
	}
	err = disputeUc.walletHandler.Release(releaseRequest)
	if err != nil {
		return err
	}
	err = disputeUc.orderRepo.UpdateOrderStatus(order.ID, domain.StatusCompleted)
	if err != nil {
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

func (disputeUc *DefaultDisputeUsecase) RejectDispute(disputeID string) error {
	err :=  disputeUc.disputeRepo.UpdateDisputeStatus(disputeID, domain.DisputeRejected)
	if err != nil {
		return err
	}
	dispute, err := disputeUc.disputeRepo.GetDisputeByID(disputeID)
	if err != nil {
		return err
	}
	order, err := disputeUc.orderRepo.GetOrderByID(dispute.OrderID)
	if err != nil {
		return err
	}
	bankDetailID := order.BankDetailID
	bankDetail, err := disputeUc.bankDetailRepo.GetBankDetailByID(bankDetailID)
	if err != nil {
		return err
	}
	releaseRequest := walletRequest.ReleaseRequest{
		TraderID: bankDetail.TraderID,
		MerchantID: order.MerchantInfo.MerchantID,
		OrderID: order.ID,
		RewardPercent: 1,
		PlatformFee: 1,
	}
	err = disputeUc.walletHandler.Release(releaseRequest)
	if err != nil {
		return err
	}
	err = disputeUc.orderRepo.UpdateOrderStatus(order.ID, domain.StatusCanceled)
	if err != nil {
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

func (disputeUc *DefaultDisputeUsecase) FreezeDispute(disputeID string) error {
	dispute, err := disputeUc.disputeRepo.GetDisputeByID(disputeID)
	if err != nil {
		return err
	}
	if dispute.Status != domain.DisputeOpened {
		return status.Error(codes.FailedPrecondition, "dispute is not opened yet")
	}
	err =  disputeUc.disputeRepo.UpdateDisputeStatus(disputeID, domain.DisputeFreezed)
	if err != nil {
		return err
	}
	return nil
}

func (disputeUc *DefaultDisputeUsecase) GetDisputeByID(disputeID string) (*domain.Dispute, error) {
	return disputeUc.disputeRepo.GetDisputeByID(disputeID)
}

func (disputeUc *DefaultDisputeUsecase) GetDisputeByOrderID(orderID string) (*domain.Dispute, error) {
	return disputeUc.disputeRepo.GetDisputeByOrderID(orderID)
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

func (disputeUc *DefaultDisputeUsecase) GetOrderDisputes(page, limit int64, status string) ([]*domain.Dispute, int64, error) {
	return disputeUc.disputeRepo.GetOrderDisputes(page, limit, status)
}