package usecase

import (
	"context"
	"log"
	"log/slog"
	"time"

	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	publisher "github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (uc *DefaultOrderUsecase) CancelOrder(orderID string) error {
	// Find exact order
	order, err := uc.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != domain.StatusPending && order.Status != domain.StatusDisputeCreated{
		return domain.ErrCancelOrder
	}

	if order.Type == domain.TypePayIn {
		return uc.processPayInCancel(order)
	}else if order.Type == domain.TypePayOut {
		return uc.processPayOutCancel(order)
	}

	return status.Errorf(codes.Internal, "failed to cancel order: unknown order type")
}

func (uc *DefaultOrderUsecase) processPayInCancel(order *domain.Order) error {
	orderID := order.ID
	op := &OrderOperation{
        OrderID:   orderID,
        Operation: "cancel",
        OldStatus: order.Status,
        NewStatus: domain.StatusCanceled,
        WalletOp: &WalletOperation{
            Type: "release",
            Request: walletRequest.ReleaseRequest{
                TraderID:      order.RequisiteDetails.TraderID,
                MerchantID:    order.MerchantInfo.MerchantID,
                OrderID:       order.ID,
                RewardPercent: 1, // При отмене не даем вознаграждение !!!!!!!!!!!!!!!!!!!!!!!!!!!!
                PlatformFee:   1, // При отмене не берем комиссию
            },
        },
		CreatedAt: time.Now(),
	}

	if err := uc.ProcessOrderOperation(context.Background(), op); err != nil {
		return err
	}

	go func(event publisher.OrderEvent){
		if err := uc.Publisher.PublishOrder(event); err != nil {
			slog.Error("failed to publish kafka order event", "stage", "cancelling", "error", err.Error())
		}
	}(publisher.OrderEvent{
		OrderID: order.ID,
		TraderID: order.RequisiteDetails.TraderID,
		Status: "⛔️Отмена сделки",
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
			string(domain.StatusCanceled),
			0, 0, 0,
		)
	}

	// ✅ ЗАПИСЬ МЕТРИКИ ОТМЕНЕННОГО ЗАКАЗА
	uc.recordOrderCanceledMetrics(order)
	
	return nil
}

func (uc *DefaultOrderUsecase) processPayOutCancel(order *domain.Order) error {
	orderID := order.ID
	op := &OrderOperation{
        OrderID:   orderID,
        Operation: "cancel",
        OldStatus: order.Status,
        NewStatus: domain.StatusCanceled,
        WalletOp: &WalletOperation{
            Type: "release",
            Request: walletRequest.ReleaseRequest{
                TraderID:      order.RequisiteDetails.TraderID,
                MerchantID:    order.MerchantInfo.MerchantID,
                OrderID:       order.ID,
                RewardPercent: 0, // При отмене не даем вознаграждение !!!!!!!!!!!!!!!!!!!!!!!!!!!!
                PlatformFee:   0, // При отмене не берем комиссию
            },
        },
		CreatedAt: time.Now(),
	}

	if err := uc.ProcessOrderOperation(context.Background(), op); err != nil {
		return err
	}

	go func(event publisher.OrderEvent){
		if err := uc.Publisher.PublishOrder(event); err != nil {
			slog.Error("failed to publish kafka order event", "stage", "cancelling", "error", err.Error())
		}
	}(publisher.OrderEvent{
		OrderID: order.ID,
		TraderID: order.RequisiteDetails.TraderID,
		Status: "⛔️Отмена выплаты",
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
			string(domain.StatusCanceled),
			0, 0, 0,
		)
	}

	return nil
}

func (uc *DefaultOrderUsecase) CancelExpiredOrders(ctx context.Context) error {
	orders, err := uc.FindExpiredOrders()
	if err != nil {
		return nil
	}

	for _, order := range orders {
		err = uc.CancelOrder(order.ID)
		if err != nil {
			log.Printf("Failed to cancel order %s to timeout! Error: %v\n", order.ID, err)
		}

		log.Printf("Order %s canceled due to timeout!\n", order.ID)
	}

	return nil
}