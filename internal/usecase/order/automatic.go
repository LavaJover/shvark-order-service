package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"math"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	publisher "github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
	"github.com/google/uuid"
)

type AutomaticPaymentRequest struct {
	Group         string
	Amount        float64
	PaymentSystem string
	Direction 	  string
	Methods       []string
	ReceivedAt    int64
	Text          string
	Metadata      map[string]string
}

func (uc *DefaultOrderUsecase) ProcessAutomaticPayment(ctx context.Context, req *AutomaticPaymentRequest) (*domain.AutomaticPaymentResult, error) {
    startTime := time.Now()
    
    log.Printf("🤖 [AUTOMATIC] Starting payment processing: device=%s, amount=%.2f, payment_system=%s", 
        req.Group, req.Amount, req.PaymentSystem)
    
    // Создаем доменный объект лога
    automaticLog := &domain.AutomaticLog{
        ID:            uuid.New().String(),
        DeviceID:      req.Group,
        Amount:        req.Amount,
        PaymentSystem: req.PaymentSystem,
        Direction:     req.Direction,
        Methods:       req.Methods,
        ReceivedAt:    time.Unix(req.ReceivedAt, 0),
        Text:          req.Text,
        CreatedAt:     time.Now(),
    }
    
    // 1. Поиск подходящих сделок
    log.Printf("🔍 [AUTOMATIC] Searching for matching orders: device=%s, amount=%.2f", req.Group, req.Amount)
    
    orders, err := uc.findMatchingOrders(ctx, req)
    if err != nil {
        log.Printf("❌ [AUTOMATIC] Error searching orders: %v", err)
        
        automaticLog.Action = "search_error"
        automaticLog.Success = false
        automaticLog.OrdersFound = 0
        automaticLog.ErrorMessage = err.Error()
        automaticLog.ProcessingTime = time.Since(startTime).Milliseconds()
        
        // Сохраняем лог (ошибки игнорируем, чтобы не блокировать основной процесс)
        if saveErr := uc.OrderRepo.SaveAutomaticLog(ctx, automaticLog); saveErr != nil {
            log.Printf("⚠️  [AUTOMATIC] Failed to save log: %v", saveErr)
        }
        
        return nil, fmt.Errorf("failed to find matching orders: %w", err)
    }
    
    automaticLog.OrdersFound = len(orders)
    
    if len(orders) == 0 {
        log.Printf("⚠️  [AUTOMATIC] No matching orders found: device=%s, amount=%.2f", req.Group, req.Amount)
        
        automaticLog.Action = "not_found"
        automaticLog.Success = false
        automaticLog.ProcessingTime = time.Since(startTime).Milliseconds()
        
        if saveErr := uc.OrderRepo.SaveAutomaticLog(ctx, automaticLog); saveErr != nil {
            log.Printf("⚠️  [AUTOMATIC] Failed to save log: %v", saveErr)
        }
        
        return &domain.AutomaticPaymentResult{
            Action:  "not_found",
            Message: "no matching orders found",
        }, nil
    }
    
    log.Printf("✅ [AUTOMATIC] Found %d matching order(s)", len(orders))
    
    // Логируем каждый найденный заказ
    for i, order := range orders {
        log.Printf("   [%d] OrderID=%s, Amount=%.2f, Status=%s, TraderID=%s, BankName=%s", 
            i+1, order.ID, order.AmountInfo.AmountFiat, order.Status, 
            order.RequisiteDetails.TraderID, order.RequisiteDetails.BankName)
    }
    
    // 2. Обработка найденных сделок
    results := make([]domain.OrderProcessingResult, 0, len(orders))
    successCount := 0
    
    for _, order := range orders {
        log.Printf("🔄 [AUTOMATIC] Processing order %s", order.ID)
        
        result, err := uc.processSingleOrder(ctx, order, req)
        if err != nil {
            log.Printf("❌ [AUTOMATIC] Failed to process order %s: %v", order.ID, err)
            automaticLog.ErrorMessage = err.Error()
            continue
        }
        
        if result.Success {
            successCount++
            log.Printf("✅ [AUTOMATIC] Order %s processed successfully", order.ID)
            
            // Обновляем лог первым успешным заказом
            if automaticLog.OrderID == "" {
                automaticLog.OrderID = order.ID
                automaticLog.TraderID = order.RequisiteDetails.TraderID
                automaticLog.BankName = order.RequisiteDetails.BankName
                automaticLog.CardNumber = order.RequisiteDetails.CardNumber
            }
        } else {
            log.Printf("⚠️  [AUTOMATIC] Order %s: %s", order.ID, result.Action)
        }
        
        results = append(results, result)
    }
    
    // Финализируем лог
    automaticLog.ProcessingTime = time.Since(startTime).Milliseconds()
    automaticLog.Success = successCount > 0
    
    if successCount > 0 {
        automaticLog.Action = "approved"
    } else {
        automaticLog.Action = "failed"
    }
    
    if saveErr := uc.OrderRepo.SaveAutomaticLog(ctx, automaticLog); saveErr != nil {
        log.Printf("⚠️  [AUTOMATIC] Failed to save log: %v", saveErr)
    }
    
    log.Printf("🏁 [AUTOMATIC] Processing completed: success=%d/%d, time=%dms", 
        successCount, len(orders), automaticLog.ProcessingTime)
    
    return &domain.AutomaticPaymentResult{
        Action:  "processed",
        Results: results,
    }, nil
}


func (uc *DefaultOrderUsecase) findMatchingOrders(ctx context.Context, req *AutomaticPaymentRequest) ([]*domain.Order, error) {
	// Поиск по device_id (group) и статусу PENDING
	orders, err := uc.OrderRepo.FindPendingOrdersByDeviceID(req.Group)
	if err != nil {
		return nil, err
	}

	// Фильтрация по сумме (с допуском ±1%)
	var matchingOrders []*domain.Order
	for _, order := range orders {
		if uc.isAmountMatching(order.AmountInfo.AmountFiat, req.Amount) {
			matchingOrders = append(matchingOrders, order)
		}
	}

	return matchingOrders, nil
}

func (uc *DefaultOrderUsecase) isAmountMatching(orderAmount, paymentAmount float64) bool {
	// Допуск 1% для учета возможных расхождений
	diff := math.Abs((orderAmount - paymentAmount))
	allowedDiff := orderAmount * 0
	return diff <= allowedDiff
}

func (uc *DefaultOrderUsecase) processSingleOrder(ctx context.Context, order *domain.Order, req *AutomaticPaymentRequest) (domain.OrderProcessingResult, error) {
	// Проверяем, не обработана ли уже сделка
	if order.Status != domain.StatusPending {
		return domain.OrderProcessingResult{
			OrderID: order.ID,
			Action:  "already_processed",
			Success: false,
		}, nil
	}

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

	// Создаем операцию для подтверждения сделки
	op := &OrderOperation{
		OrderID:   order.ID,
		Operation: "auto_approve",
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
		// Metadata: map[string]interface{}{
		// 	"automatic_payment": true,
		// 	"received_at":       req.ReceivedAt,
		// 	"payment_system":    req.PaymentSystem,
		// 	"source":            "sms_parser",
		// },
		CreatedAt: time.Now(),
	}

	// Выполняем операцию
	if err := uc.ProcessOrderOperation(ctx, op); err != nil {
		return domain.OrderProcessingResult{
			OrderID: order.ID,
			Action:  "failed",
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Публикуем событие
	go uc.publishAutomaticApprovalEvent(order, req)

	if order.CallbackUrl != "" {
		notifier.SendCallback(
			order.CallbackUrl,
			order.MerchantInfo.MerchantOrderID,
			string(domain.StatusCompleted),
			0, 0, 0,
		)
	}

	return domain.OrderProcessingResult{
		OrderID: order.ID,
		Action:  "approved",
		Success: true,
	}, nil
}

func (uc *DefaultOrderUsecase) generatePaymentHash(req *AutomaticPaymentRequest) string {
	// Создаем уникальный хэш для уведомления чтобы избежать дублирующей обработки
	data := fmt.Sprintf("%s_%.2f_%d", req.Group, req.Amount, req.ReceivedAt)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (uc *DefaultOrderUsecase) ensureIdempotency(ctx context.Context, orderID string, paymentHash string) (bool, error) {
	processed, err := uc.OrderRepo.CheckDuplicatePayment(ctx, orderID, paymentHash)
	if err != nil {
		return false, err
	}
	return processed, nil
}

func (uc *DefaultOrderUsecase) publishAutomaticApprovalEvent(order *domain.Order, req *AutomaticPaymentRequest) {
	event := publisher.OrderEvent{
		OrderID:     order.ID,
		TraderID:    order.RequisiteDetails.TraderID,
		Status:      "✅ Автоматически закрыта",
		AmountFiat:  order.AmountInfo.AmountFiat,
		Currency:    order.AmountInfo.Currency,
		BankName:    order.RequisiteDetails.BankName,
		Phone:       order.RequisiteDetails.Phone,
		CardNumber:  order.RequisiteDetails.CardNumber,
		Owner:       order.RequisiteDetails.Owner,
		// Metadata: map[string]interface{}{
		// 	"automatic":    true,
		// 	"payment_system": req.PaymentSystem,
		// 	"source":       "sms_parser",
		// },
	}
	
	if err := uc.Publisher.PublishOrder(event); err != nil {
		slog.Error("failed to publish automatic approval event", 
			"order_id", order.ID, 
			"error", err.Error())
	}
}