package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	publisher "github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
)

////////////////////// Advanced Safe Order operations //////////////////////////

// OrderOperation - описание операции со сделкой
type OrderOperation struct {
    OrderID     string                    `json:"order_id"`
    Operation   string                    `json:"operation"` // "create", "approve", "cancel"
    OldStatus   domain.OrderStatus        `json:"old_status"`
    NewStatus   domain.OrderStatus        `json:"new_status"`
    WalletOp    *WalletOperation         `json:"wallet_op,omitempty"`
    CreatedAt   time.Time                `json:"created_at"`
}

type WalletOperation struct {
    Type    string  `json:"type"` // "freeze", "release"
    Request interface{} `json:"request"`
}

// OrderTransactionState - состояние транзакции операции
type OrderTransactionState struct {
    OrderID         string    `json:"order_id"`
    Operation       string    `json:"operation"`
    StatusChanged   bool      `json:"status_changed"`
    WalletProcessed bool      `json:"wallet_processed"`
    EventPublished  bool      `json:"event_published"`
    CallbackSent    bool      `json:"callback_sent"`
    CreatedAt       time.Time `json:"created_at"`
    CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

///////////////////////// Базовая транзакционная функция //////////////////////////

// ProcessOrderOperation - базовая функция для всех операций со сделками
func (uc *DefaultOrderUsecase) ProcessOrderOperation(ctx context.Context, op *OrderOperation) error {
    // 1. КРИТИЧНО: Атомарно меняем статус и обрабатываем кошелек
    if err := uc.processCriticalOperations(ctx, op); err != nil {
        return fmt.Errorf("critical operations failed: %w", err)
    }

    // 2. НЕКРИТИЧНО: Асинхронно публикуем событие и отправляем callback
    // if err := uc.scheduleNonCriticalOperations(op); err != nil {
    //     log.Printf("Failed to schedule non-critical operations for order %s: %v", op.OrderID, err)
    //     // НЕ возвращаем ошибку - критичные операции уже выполнены
    // }

    return nil
}

// processCriticalOperations - синхронная обработка критичных операций
func (uc *DefaultOrderUsecase) processCriticalOperations(ctx context.Context, op *OrderOperation) error {
    var walletFunc func() error
    if op.WalletOp != nil {
        walletFunc = func() error {
            return uc.processWalletOperation(op.WalletOp)
        }
    }

    return uc.OrderRepo.ProcessOrderCriticalOperation(
        op.OrderID, 
        op.NewStatus, 
        op.Operation, // передаем тип операции
        walletFunc,
    )
}

// processWalletOperation - обработка операций с кошельком
func (uc *DefaultOrderUsecase) processWalletOperation(walletOp *WalletOperation) error {
    switch walletOp.Type {
    case "freeze":
        req := walletOp.Request.(walletRequest.FreezeRequest)
        return uc.WalletHandler.Freeze(req.TraderID, req.OrderID, req.Amount)
    case "release":
        req := walletOp.Request.(walletRequest.ReleaseRequest)
        return uc.WalletHandler.Release(req)
    default:
        return fmt.Errorf("unknown wallet operation: %s", walletOp.Type)
    }
}

func (uc *DefaultOrderUsecase) cancelOrderDueToFreezeFailure(order *domain.Order, freezeErr error) {
    slog.Error("Freeze failed after order creation, canceling order", "order_id", order.ID, "error", freezeErr)
    
    // Пытаемся отменить заказ
    if err := uc.CancelOrder(order.ID); err != nil {
        slog.Error("Failed to cancel order after freeze failure", "order_id", order.ID, "error", err)
    }
    
    // Отправляем колбэк об ошибке
    if order.CallbackUrl != "" {
        notifier.SendCallback(
            order.CallbackUrl,
            order.MerchantInfo.MerchantOrderID,
            string(domain.StatusFailed),
            0, 0, 0,
        )
    }
}

func (uc *DefaultOrderUsecase) sendOrderNotifications(order *domain.Order, bankDetail *domain.BankDetail) {
    // Publish to Kafka асинхронно
    go func(event publisher.OrderEvent) {
        if err := uc.Publisher.PublishOrder(event); err != nil {
            slog.Error("failed to publish OrderEvent:created", "error", err.Error())
        }
    }(publisher.OrderEvent{
        OrderID:   order.ID,
        TraderID:  order.RequisiteDetails.TraderID,
        Status:    "🔥Новая сделка",
        AmountFiat: order.AmountInfo.AmountFiat,
        Currency:  order.AmountInfo.Currency,
        BankName:  order.RequisiteDetails.BankName,
        Phone:     order.RequisiteDetails.Phone,
        CardNumber: order.RequisiteDetails.CardNumber,
        Owner:     order.RequisiteDetails.Owner,
    })

    if order.CallbackUrl != "" {
        notifier.SendCallback(
            order.CallbackUrl,
            order.MerchantInfo.MerchantOrderID,
            string(domain.StatusPending),
            0, 0, 0,
        )
    }
}