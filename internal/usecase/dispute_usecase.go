package usecase

import (
	"log"

	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DefaultDisputeUsecase struct {
	disputeRepo domain.DisputeRepository
	walletHandler *handlers.HTTPWalletHandler
	orderRepo domain.OrderRepository
	trafficRepo domain.TrafficRepository
	kafkaPublisher *kafka.KafkaPublisher
}

func NewDefaultDisputeUsecase(
	disputeRepo domain.DisputeRepository,
	walletHandler *handlers.HTTPWalletHandler,
	orderRepo domain.OrderRepository,
	trafficRepo domain.TrafficRepository,
	kafkaPublisher *kafka.KafkaPublisher,
	) *DefaultDisputeUsecase {
	return &DefaultDisputeUsecase{
		disputeRepo: disputeRepo,
		walletHandler: walletHandler,
		orderRepo: orderRepo,
		trafficRepo: trafficRepo,
		kafkaPublisher: kafkaPublisher,
	}
}

// Диспут открыт -> запись в БД со статусом DISPUTE_OPENED
// AutoAcceptAt -> в данное время система автоматически одобрит диспут
func (disputeUc *DefaultDisputeUsecase) CreateDispute(dispute *domain.Dispute) error {
	order, err := disputeUc.orderRepo.GetOrderByID(dispute.OrderID)
	if err != nil {
		return err
	}
	// Если сделка не ушла в отмену, то диспут не открыть
	if order.Status != domain.StatusCanceled {
		return status.Error(codes.FailedPrecondition, "order is not even canceled")
	}
	dispute.DisputeCryptoRate = order.CryptoRubRate
	dispute.DisputeAmountCrypto = dispute.DisputeAmountFiat / dispute.DisputeCryptoRate
	err = disputeUc.disputeRepo.CreateDispute(dispute)
	if err != nil {
		return err
	}
	disputeUc.kafkaPublisher.PublishDispute(kafka.DisputeEvent{
		DisputeID: dispute.ID,
		OrderID: dispute.OrderID,
		TraderID: order.BankDetail.TraderID,
		OrderAmountFiat: order.AmountFiat,
		DisputeAmountFiat: dispute.DisputeAmountFiat,
		ProofUrl: dispute.ProofUrl,
		Reason: dispute.Reason,
		Status: "🆘Открыт диспут",
		BankName: order.BankDetail.BankName,
		Phone: order.BankDetail.Phone,
		CardNumber: order.BankDetail.CardNumber,
		Owner: order.BankDetail.Owner,
	})
	err = disputeUc.walletHandler.Freeze(order.BankDetail.TraderID, dispute.OrderID, dispute.DisputeAmountCrypto)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	err = disputeUc.orderRepo.UpdateOrderStatus(order.ID, domain.OrderStatus(domain.DisputeOpened))
	if err != nil {
		return err
	}
	if order.CallbackURL != "" {
		notifier.SendCallback(
			order.CallbackURL,
			order.MerchantOrderID,
			string(domain.DisputeOpened),
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
	traffic, err := disputeUc.trafficRepo.GetTrafficByTraderMerchant(order.BankDetail.TraderID, order.MerchantID)
	if err != nil {
		return err
	}
	err = disputeUc.walletHandler.Release(order.BankDetail.TraderID, order.MerchantID, order.ID, traffic.TraderRewardPercent, traffic.PlatformFee)
	if err != nil {
		return err
	}
	err = disputeUc.orderRepo.UpdateOrderStatus(order.ID, domain.StatusCompleted)
	if err != nil {
		return err
	}
	if order.CallbackURL != "" {
		notifier.SendCallback(
			order.CallbackURL,
			order.MerchantOrderID,
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
	err = disputeUc.walletHandler.Release(order.BankDetail.TraderID, order.MerchantID, order.ID, 1., 0.)
	if err != nil {
		return err
	}
	err = disputeUc.orderRepo.UpdateOrderStatus(order.ID, domain.StatusCanceled)
	if err != nil {
		return err
	}
	if order.CallbackURL != "" {
		notifier.SendCallback(
			order.CallbackURL,
			order.MerchantOrderID,
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