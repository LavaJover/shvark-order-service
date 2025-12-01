package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
)

////////////////////// Advanced Safe Dispute operations //////////////////////////

// DisputeOperation - описание операции с диспутом
type DisputeOperation struct {
	OrderID		string								`json:"order_id"`
    DisputeID   string                   			`json:"dispute_id"`
    Operation   string                    			`json:"operation"` // "create", "approve", "cancel", "freeze"
	OldOrderStatus domain.OrderStatus				`json:"old_order_status"`
	NewOrderStatus domain.OrderStatus				`json:"new_order_status"`
    OldDisputeStatus   domain.DisputeStatus        	`json:"old_status"`
    NewDisputeStatus   domain.DisputeStatus        	`json:"new_status"`
	NewOrderAmountFiat float64
	NewOrderAmountCrypto float64
	NewOrderAmountCryptoRate float64
    WalletOp    *WalletOperation         			`json:"wallet_op,omitempty"`
    CreatedAt   time.Time                			`json:"created_at"`
}

type WalletOperation struct {
    Type    string  `json:"type"` // "freeze", "release"
    Request interface{} `json:"request"`
}

///////////////////////// Базовая транзакционная функция //////////////////////////

// ProcessOrderOperation - базовая функция для всех операций со сделками
func (disputeUc *DefaultDisputeUsecase) ProcessDisputeOperation(ctx context.Context, op *DisputeOperation) error {
    // 1. КРИТИЧНО: Атомарно меняем статус и обрабатываем кошелек
    if err := disputeUc.processCriticalOperations(ctx, op); err != nil {
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
func (disputeUc *DefaultDisputeUsecase) processCriticalOperations(ctx context.Context, op *DisputeOperation) error {
    var walletFunc func() error
    if op.WalletOp != nil {
        walletFunc = func() error {
            return disputeUc.processWalletOperation(op.WalletOp)
        }
    }

    return disputeUc.disputeRepo.ProcessDisputeCriticalOperation(
        op.DisputeID, 
		op.OrderID,
        op.NewDisputeStatus,
		op.NewOrderStatus,
		op.NewOrderAmountFiat, op.NewOrderAmountCrypto, op.NewOrderAmountCryptoRate,
        op.Operation, // передаем тип операции
        walletFunc,
    )
}

// processWalletOperation - обработка операций с кошельком
func (disputeUc *DefaultDisputeUsecase) processWalletOperation(walletOp *WalletOperation) error {
    switch walletOp.Type {
    case "freeze":
        req := walletOp.Request.(walletRequest.FreezeRequest)
        return disputeUc.walletHandler.Freeze(req.TraderID, req.OrderID, req.Amount)
    case "release":
        req := walletOp.Request.(walletRequest.ReleaseRequest)
        return disputeUc.walletHandler.Release(req)
    default:
        return fmt.Errorf("unknown wallet operation: %s", walletOp.Type)
    }
}