package usecase

import (
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	publisher "github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	disputedto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/dispute"
)

type DisputeUsecase interface {
	CreateDispute(input *disputedto.CreateDisputeInput) error
	AcceptDispute(disputeID string) error
	RejectDispute(disputeID string) error
	FreezeDispute(disputeID string) error
	GetDisputeByID(disputeID string) (*domain.Dispute, error)
	GetDisputeByOrderID(orderID string) (*domain.Dispute, error)
	AcceptExpiredDisputes() error
	GetOrderDisputes(input *disputedto.GetOrderDisputesInput) (*disputedto.GetOrderDisputesOutput, error)
}

type DefaultDisputeUsecase struct {
	disputeRepo domain.DisputeRepository
	walletHandler *handlers.HTTPWalletHandler
	orderRepo domain.OrderRepository
	trafficRepo domain.TrafficRepository
	kafkaPublisher *publisher.KafkaPublisher
	teamRelationsUsecase usecase.TeamRelationsUsecase
	bankDetailRepo domain.BankDetailRepository
	TrafficUsecase usecase.TrafficUsecase
}

func NewDefaultDisputeUsecase(
	disputeRepo domain.DisputeRepository,
	walletHandler *handlers.HTTPWalletHandler,
	orderRepo domain.OrderRepository,
	trafficRepo domain.TrafficRepository,
	kafkaPublisher *publisher.KafkaPublisher,
	teamRelationsUsecase usecase.TeamRelationsUsecase,
	bankDetailRepo domain.BankDetailRepository,
	trafficUsecase usecase.TrafficUsecase,
	) *DefaultDisputeUsecase {
	return &DefaultDisputeUsecase{
		disputeRepo: disputeRepo,
		walletHandler: walletHandler,
		orderRepo: orderRepo,
		trafficRepo: trafficRepo,
		kafkaPublisher: kafkaPublisher,
		teamRelationsUsecase: teamRelationsUsecase,
		bankDetailRepo: bankDetailRepo,
		TrafficUsecase: trafficUsecase,
	}
}