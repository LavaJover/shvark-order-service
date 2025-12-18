package usecase

import (
	"context"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/metrics"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	orderdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/order"
)

type OrderUsecase interface {
	CreatePayInOrder(input *orderdto.CreatePayInOrderInput) (*orderdto.OrderOutput, error)
    CreatePayInOrderAtomic(input *orderdto.CreatePayInOrderInput) (*orderdto.OrderOutput, error) // Добавляем новый метод
    CreatePayOutOrder(input *orderdto.CreatePayOutOrderInput) (*orderdto.OrderOutput, error)

    AcceptOrder(orderID string) error
	ApproveOrder(orderID string) error
	CancelOrder(orderID string) error
    CancelExpiredOrders(context.Context) error

	GetOrderByID(orderID string) (*domain.Order, error)
	GetOrderByMerchantOrderID(merchantOrderID string) (*domain.Order, error)
	GetOrdersByTraderID(
		orderID string, page, 
		limit int64, sortBy, 
		sortOrder string, 
		filters domain.OrderFilters,
		) ([]*orderdto.OrderOutput, int64, error)
	FindExpiredOrders() ([]*domain.Order, error)
    GetOrders(filter domain.Filter, sortField string, page, size int) ([]*domain.Order, int64, error)
	GetAllOrders(input *orderdto.GetAllOrdersInput) (*orderdto.GetAllOrdersOutput, error)
    GetOrderStatistics(traderID string, dateFrom, dateTo time.Time) (*domain.OrderStatistics, error)

	ProcessAutomaticPayment(ctx context.Context, req *AutomaticPaymentRequest) (*domain.AutomaticPaymentResult, error)
}

type DefaultOrderUsecase struct {
	OrderRepo 			domain.OrderRepository
	WalletHandler   	*handlers.HTTPWalletHandler
	TrafficUsecase  	usecase.TrafficUsecase
	BankDetailUsecase 	usecase.BankDetailUsecase
	TeamRelationsUsecase usecase.TeamRelationsUsecase
	Publisher 			*publisher.KafkaPublisher
	Metrics				*metrics.OrderMetrics	
}

func NewDefaultOrderUsecase(
	orderRepo domain.OrderRepository, 
	walletHandler *handlers.HTTPWalletHandler,
	trafficUsecase usecase.TrafficUsecase,
	bankDetailUsecase usecase.BankDetailUsecase,
	kafkaPublisher *publisher.KafkaPublisher,
	teamRelationsUsecase usecase.TeamRelationsUsecase,
	orderMetrics *metrics.OrderMetrics) *DefaultOrderUsecase {

	return &DefaultOrderUsecase{
		OrderRepo: orderRepo,
		WalletHandler: walletHandler,
		TrafficUsecase: trafficUsecase,
		BankDetailUsecase: bankDetailUsecase,
		Publisher: kafkaPublisher,
		TeamRelationsUsecase: teamRelationsUsecase,
		Metrics: orderMetrics,
	}
}
