package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/dto"
	bankdetaildto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/bank_detail"
	orderdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/order"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderUsecase interface {
	CreateOrder(input *orderdto.CreateOrderInput) (*orderdto.OrderOutput, error)
	GetOrderByID(orderID string) (*orderdto.OrderOutput, error)
	GetOrderByMerchantOrderID(merchantOrderID string) (*orderdto.OrderOutput, error)
	GetOrdersByTraderID(
		orderID string, page, 
		limit int64, sortBy, 
		sortOrder string, 
		filters domain.OrderFilters,
		) ([]*orderdto.OrderOutput, int64, error)
	FindExpiredOrders() ([]*domain.Order, error)
	CancelExpiredOrders(context.Context) error
	ApproveOrder(orderID string) error
	CancelOrder(orderID string) error
	GetOrderStatistics(traderID string, dateFrom, dateTo time.Time) (*domain.OrderStatistics, error)

	GetOrders(filter domain.Filter, sortField string, page, size int) ([]*domain.Order, int64, error)

	GetAllOrders(input *orderdto.GetAllOrdersInput) (*orderdto.GetAllOrdersOutput, error)
}

type DefaultOrderUsecase struct {
	OrderRepo 			domain.OrderRepository
	WalletHandler   	*handlers.HTTPWalletHandler
	TrafficUsecase  	domain.TrafficUsecase
	BankDetailUsecase 	BankDetailUsecase
	TeamRelationsUsecase TeamRelationsUsecase
	mqPub				domain.PublisherPort
	mqSub 				domain.SubscriberPort
}

func NewDefaultOrderUsecase(
	orderRepo domain.OrderRepository, 
	walletHandler *handlers.HTTPWalletHandler,
	trafficUsecase domain.TrafficUsecase,
	bankDetailUsecase BankDetailUsecase,
	teamRelationsUsecase TeamRelationsUsecase,
	pub domain.PublisherPort,
	sub domain.SubscriberPort) *DefaultOrderUsecase {

	return &DefaultOrderUsecase{
		OrderRepo: orderRepo,
		WalletHandler: walletHandler,
		TrafficUsecase: trafficUsecase,
		BankDetailUsecase: bankDetailUsecase,
		TeamRelationsUsecase: teamRelationsUsecase,
		mqPub: pub,
		mqSub: sub,
	}
}

func (uc *DefaultOrderUsecase) PickBestBankDetail(bankDetails []*domain.BankDetail, merchantID string) (*domain.BankDetail, error) {
	if len(bankDetails) == 0 {
		return nil, fmt.Errorf("no available bank details provided to pick the best")
	}
	type Trader struct {
		TraderID 		string
		Priority 		float64
		BankDetailIndex int
	}
	var traders []*Trader
	totalPriority := 0.0

	for i, bankDetail := range bankDetails {
		traderID := bankDetail.TraderID
		traffic, err := uc.TrafficUsecase.GetTrafficByTraderMerchant(traderID, merchantID)
		if err != nil {
			fmt.Println("Error while picking trader: " + err.Error())
			return nil, err
		}
		traders = append(traders, &Trader{
			TraderID: traffic.TraderID,
			Priority: traffic.TraderPriority,
			BankDetailIndex: i,
		})
		totalPriority += traffic.TraderPriority
	}

	// [0, totalPriority]
	rand.Seed(time.Now().UnixNano())
	r := rand.Float64() * totalPriority

	// random shuffle of array
	rand.Shuffle(len(traders), func(i, j int) {
		traders[i], traders[j] = traders[j], traders[i]
	})

	// pick trader regarding weight
	accumulated := 0.0
	for _, trader := range traders {
		accumulated += trader.Priority
		if r <= accumulated {
			return bankDetails[trader.BankDetailIndex], nil
		}
	}

	return bankDetails[traders[len(traders)-1].BankDetailIndex], nil
}

func (uc *DefaultOrderUsecase) FilterByTraffic(bankDetails []*domain.BankDetail, merchantID string) ([]*domain.BankDetail, error) {
	result := make([]*domain.BankDetail, 0)
	for _, bankDetail := range bankDetails {
		traffic, err := uc.TrafficUsecase.GetTrafficByTraderMerchant(bankDetail.TraderID, merchantID)
		if err != nil {
			continue
		}
		if traffic.Enabled {
			result = append(result, bankDetail)
		}
	}

	return result, nil
}

// FilterByTraderBalanceOptimal - оптимизированная версия с пакетным запросом
func (uc *DefaultOrderUsecase) FilterByTraderBalanceOptimal(bankDetails []*domain.BankDetail, amountCrypto float64) ([]*domain.BankDetail, error) {
	startTime := time.Now()
	defer func() {
		log.Printf("FilterByTraderBalanceOptimal took %v", time.Since(startTime))
	}()

	if len(bankDetails) == 0 {
		return []*domain.BankDetail{}, nil
	}

	// Собираем уникальные traderIDs
	traderIDMap := make(map[string]bool)
	for _, bankDetail := range bankDetails {
		traderIDMap[bankDetail.TraderID] = true
	}

	traderIDs := make([]string, 0, len(traderIDMap))
	for traderID := range traderIDMap {
		traderIDs = append(traderIDs, traderID)
	}

	// Получаем балансы одним запросом
	balances, err := uc.WalletHandler.GetTraderBalancesBatch(traderIDs)
	if err != nil {
		fmt.Println(err.Error())
		return nil, fmt.Errorf("failed to get trader balances: %w", err)
	}

	// Фильтруем банковские реквизиты
	result := make([]*domain.BankDetail, 0, len(bankDetails))
	validCount := 0

	for _, bankDetail := range bankDetails {
		balance, exists := balances[bankDetail.TraderID]
		if !exists {
			log.Printf("Trader %s not found in balances", bankDetail.TraderID)
			continue
		}

		if balance >= amountCrypto {
			result = append(result, bankDetail)
			validCount++
		} else {
			log.Printf("Trader %s insufficient balance: %f < %f", 
				bankDetail.TraderID, balance, amountCrypto)
		}
	}

	log.Printf("FilterByTraderBalance: %d/%d traders have sufficient balance", 
		validCount, len(bankDetails))

	return result, nil
}

func (uc *DefaultOrderUsecase)FilterByEqualAmountFiat(bankDetails []*domain.BankDetail, amountFiat float64) ([]*domain.BankDetail, error) {
	// Отбросить реквизиты, на которых уже есть созданная заявка на сумму anountFiat
	result := make([]*domain.BankDetail, 0)
	for _, bankDetail := range bankDetails {
		fmt.Println("Проверка на одинаковую сумму!")
		orders, err := uc.OrderRepo.GetOrdersByBankDetailID(bankDetail.ID)
		if err != nil {
			return nil, err
		}
		skipBankDetail := false
		for _, order := range orders {
			if order.Status == domain.StatusPending && order.AmountInfo.AmountFiat == amountFiat {
				// Пропускаем данный рек, тк есть созданная заявка на такую сумму фиата
				skipBankDetail = true
				fmt.Println("Обнаружена активная заявка с такой же суммой!")
				break
			}
		}
		if !skipBankDetail {
			result = append(result, bankDetail)
		}
	}

	return result, nil
}

func (uc *DefaultOrderUsecase) FindEligibleBankDetails(input *orderdto.CreateOrderInput) ([]*domain.BankDetail, error) {
	bankDetails, err := uc.BankDetailUsecase.FindSuitableBankDetails(
		&bankdetaildto.FindSuitableBankDetailsInput{
			AmountFiat: input.AmountFiat,
			Currency: input.Currency,
			PaymentSystem: input.PaymentSystem,
			BankCode: input.BankInfo.BankCode,
			NspkCode: input.BankInfo.NspkCode,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(bankDetails) == 0 {
		log.Printf("Отсеились по статическим параметрам\n")
	}
	// 0) Filter by Traffic
	bankDetails, err = uc.FilterByTraffic(bankDetails, input.MerchantParams.MerchantID)
	if err != nil {
		return nil, err
	}
	if len(bankDetails) == 0 {
		log.Printf("Отсеились по трафику\n")
	}

	// 1) Filter by Trader Available balances
	bankDetails, err = uc.FilterByTraderBalanceOptimal(bankDetails, input.AmountCrypto)
	if err != nil {
		return nil, err
	}
	if len(bankDetails) == 0 {
		log.Printf("Отсеились по балансу трейдеров\n")
	}

	return bankDetails, nil
}

func (uc *DefaultOrderUsecase) CheckIdempotency(clientID string) error {
	orders, err := uc.OrderRepo.GetCreatedOrdersByClientID(clientID)
	if len(orders)!=0 || err != nil {
		return status.Errorf(codes.FailedPrecondition, "payment order already exists for client: %s", clientID)
	}

	return nil
}

func (uc *DefaultOrderUsecase) CreateOrder(createOrderInput *orderdto.CreateOrderInput) (*orderdto.OrderOutput, error) {
	start := time.Now()
	slog.Info("CreateOrder started")
	// check idempotency by client_id
	if createOrderInput.ClientID != "" {
		t := time.Now()
		if err := uc.CheckIdempotency(createOrderInput.ClientID); err != nil {
			return nil, err
		}
		slog.Info("CheckIdempotency done", "elapsed", time.Since(t))
	}

	// searching for eligible bank details due to order query parameters
	t := time.Now()
	bankDetails, err := uc.FindEligibleBankDetails(createOrderInput)
	if err != nil {
		return nil, status.Error(codes.NotFound, "no eligible bank detail"+err.Error())
	}
	slog.Info("FindEligibleBankDetails done", "elapsed", time.Since(t))
	if len(bankDetails) == 0 {
		log.Printf("Реквизиты для заявки не найдены!\n")
		return nil, fmt.Errorf("no available bank details")
	}
	log.Printf("Для заявки найдены доступные реквизиты!\n")

	if createOrderInput.AdvancedParams.CallbackUrl != "" {
		notifier.SendCallback(
			createOrderInput.AdvancedParams.CallbackUrl,
			createOrderInput.MerchantOrderID,
			string(domain.StatusCreated),
			0, 0, 0,
		)
	}

	// business logic to pick best bank detail
	t = time.Now()
	chosenBankDetail, err := uc.PickBestBankDetail(bankDetails, createOrderInput.MerchantID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to pick best bank detail for order")
	}
	slog.Info("PickBestBankDetail done", "elapsed", time.Since(t))

	// Get trader reward percent and save to order
	t = time.Now()
	traffic, err := uc.TrafficUsecase.GetTrafficByTraderMerchant(chosenBankDetail.TraderID, createOrderInput.MerchantID)
	if err != nil {
		return nil, err
	}
	slog.Info("GetTrafficByTraderMerchant done", "elapsed", time.Since(t))
	traderReward := traffic.TraderRewardPercent
	platformFee := traffic.PlatformFee

	order := domain.Order{
		ID:     uuid.New().String(),
		Status: domain.StatusPending,
		MerchantInfo: domain.MerchantInfo{
			MerchantID:     createOrderInput.MerchantID,
			MerchantOrderID: createOrderInput.MerchantOrderID,
			ClientID:       createOrderInput.ClientID,
		},
		AmountInfo: domain.AmountInfo{
			AmountFiat:   createOrderInput.AmountFiat,
			AmountCrypto: createOrderInput.AmountCrypto,
			CryptoRate:   createOrderInput.CryptoRate,
			Currency:     createOrderInput.Currency,
		},
		BankDetailID:  chosenBankDetail.ID,
		Type:          createOrderInput.Type,
		Recalculated:  createOrderInput.Recalculated,
		Shuffle:       createOrderInput.Shuffle,
		TraderReward:  traderReward,
		PlatformFee:   platformFee,
		CallbackUrl:   createOrderInput.CallbackUrl,
		ExpiresAt:     createOrderInput.ExpiresAt,
	}

    // КРИТИЧНО: Атомарно создаем заказ и замораживаем средства
    op := &OrderOperation{
        OrderID:   order.ID,
        Operation: "create",
        OldStatus: "",
        NewStatus: domain.StatusPending,
        WalletOp: &WalletOperation{
            Type: "freeze",
            Request: walletRequest.FreezeRequest{
                TraderID: chosenBankDetail.TraderID,
                OrderID:  order.ID,
                Amount:   createOrderInput.AmountCrypto,
            },
        },
        EventData: &publisher.OrderEvent{
            OrderID:    order.ID,
            TraderID:   chosenBankDetail.TraderID,
            Status:     "🔥Новая сделка",
            AmountFiat: order.AmountInfo.AmountFiat,
            Currency:   order.AmountInfo.Currency,
            BankName:   chosenBankDetail.BankName,
            Phone:      chosenBankDetail.Phone,
            CardNumber: chosenBankDetail.CardNumber,
            Owner:      chosenBankDetail.Owner,
        },
        CallbackData: &CallbackRequest{
            URL:             createOrderInput.CallbackUrl,
            MerchantOrderID: createOrderInput.MerchantOrderID,
            OrderID:         order.ID,
            Status:          string(domain.StatusPending),
        },
        CreatedAt: time.Now(),
    }
	if err := uc.ProcessOrderOperation(context.Background(), op); err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

	slog.Info("CreateOrder finished", "total_elapsed", time.Since(start))

	return &orderdto.OrderOutput{
		Order:     order,
		BankDetail: *chosenBankDetail,
	}, nil
}

func (uc *DefaultOrderUsecase) GetOrderByID(orderID string) (*orderdto.OrderOutput, error) {
	order, err := uc.OrderRepo.GetOrderByID(orderID)
	if err != nil {
		return nil, err
	}
	bankDetailID := order.BankDetailID
	bankDetail, err := uc.BankDetailUsecase.GetBankDetailByID(bankDetailID)
	if err != nil {
		return nil, err
	}
	return &orderdto.OrderOutput{
		Order: *order,
		BankDetail: *bankDetail,
	}, nil
}

func (uc *DefaultOrderUsecase) GetOrderByMerchantOrderID(merchantOrderID string) (*orderdto.OrderOutput, error) {
	order, err := uc.OrderRepo.GetOrderByMerchantOrderID(merchantOrderID)
	if err != nil {
		return nil, err
	}
	bankDetailID := order.BankDetailID
	bankDetail, err := uc.BankDetailUsecase.GetBankDetailByID(bankDetailID)
	if err != nil {
		return nil, err
	}
	return &orderdto.OrderOutput{
		Order: *order,
		BankDetail: *bankDetail,
	}, nil
}

func (uc *DefaultOrderUsecase) GetOrdersByTraderID(
	traderID string, page, 
	limit int64, sortBy, 
	sortOrder string,
	filters domain.OrderFilters,
) ([]*orderdto.OrderOutput, int64, error) {

	validStatuses := map[string]bool{
		string(domain.StatusCompleted): true,
		string(domain.StatusCanceled): true,
		string(domain.StatusPending): true,
		string(domain.StatusDisputeCreated): true,
	}

	for _, status := range filters.Statuses {
		if !validStatuses[status] {
			return nil, 0, fmt.Errorf("invalid status in filters")
		}
	}

	orders, total, err := uc.OrderRepo.GetOrdersByTraderID(
		traderID, 
		page, 
		limit, 
		sortBy, 
		sortOrder,
		filters,
	)
	if err != nil {
		return nil, 0, err
	}
	var orderOutputs []*orderdto.OrderOutput
	for _, order := range orders {
		bankDetailID := order.BankDetailID
		bankDetail, err := uc.BankDetailUsecase.GetBankDetailByID(bankDetailID)
		if err != nil {
			return nil, 0, err
		}
		orderOutputs = append(orderOutputs, &orderdto.OrderOutput{
			Order: *order,
			BankDetail: *bankDetail,
		})
	}

	return orderOutputs, total, nil
}

func (uc *DefaultOrderUsecase) FindExpiredOrders() ([]*domain.Order, error) {
	return uc.OrderRepo.FindExpiredOrders()
}

func (uc *DefaultOrderUsecase) CancelExpiredOrders(ctx context.Context) error {
    // Получаем данные отмененных заказов из БД (УЖЕ ПОЛНЫЕ ДАННЫЕ)
    expired, err := uc.OrderRepo.CancelExpiredOrdersBatch(ctx)
    if err != nil {
        return fmt.Errorf("failed to cancel expired orders: %w", err)
    }

    if len(expired) == 0 {
        return nil
    }

    log.Printf("Canceled %d expired orders, publishing task to worker...", len(expired))

    // Сериализуем ПОЛНЫЕ данные ExpiredOrderData, а не только ID
    payload, err := json.Marshal(expired)
    if err != nil {
        return fmt.Errorf("failed to marshal expired orders: %w", err)
    }

    // Публикуем полные данные в Kafka
    if err := uc.mqPub.Publish("orders.cancelled", domain.Message{Key: nil, Value: payload}); err != nil {
        log.Printf("failed to publish cancel task, will retry next tick: %v", err)
        // Не возвращаем ошибку, чтобы не блокировать основной процесс
    }

    return nil
}

func (uc *DefaultOrderUsecase) StartWorker(ctx context.Context) {
    // Подписываемся на топик отменённых ордеров
    msgs, err := uc.mqSub.Subscribe("orders.cancelled", "order-cancel-group")
    if err != nil {
        log.Fatalf("failed to subscribe to orders.cancelled: %v", err)
    }

    log.Println("Order cancel worker started")

    for {
        select {
        case <-ctx.Done():
            log.Println("Order cancel worker shutting down")
            return
        case m, ok := <-msgs:
            if !ok {
                log.Println("orders.cancelled channel closed")
                return
            }

            // Десериализуем ПОЛНЫЕ данные ExpiredOrderData
            var expiredOrders []dto.ExpiredOrderData
            if err := json.Unmarshal(m.Value, &expiredOrders); err != nil {
                log.Printf("invalid expired orders payload: %v", err)
                continue
            }

            if len(expiredOrders) == 0 {
                log.Printf("empty expired orders batch received")
                continue
            }

            log.Printf("Processing %d expired orders in worker", len(expiredOrders))

            // Вызываем готовый пайплайн обработки
            uc.handleExpiredOrdersPostProcessing(expiredOrders)

            log.Printf("Completed processing %d expired orders", len(expiredOrders))
        }
    }
}

// handleExpiredOrdersPostProcessing - обработка побочных эффектов с отслеживанием
func (uc *DefaultOrderUsecase) handleExpiredOrdersPostProcessing(expiredOrders []dto.ExpiredOrderData) {
    // 1. КРИТИЧЕСКИ ВАЖНО: Разморозка средств с отслеживанием
    uc.processWalletReleasesWithTracking(expiredOrders)

    // 2. Публикация событий (некритично)
    uc.publishOrderEventsWithTracking(expiredOrders)

    // 3. Callback'и (некритично)  
    uc.sendCallbacksWithTracking(expiredOrders)
}

type CallbackRequest struct {
    URL             string `json:"url"`
    MerchantOrderID string `json:"merchant_order_id"`
    OrderID         string `json:"order_id"` // Добавляем для отслеживания
    Status          string `json:"status"`
}

type CallbackResult struct {
    OrderID         string `json:"order_id"`
    MerchantOrderID string `json:"merchant_order_id"`
    Success         bool   `json:"success"`
    Error           string `json:"error,omitempty"`
}

// processWalletReleasesWithTracking - разморозка с детальным отслеживанием
func (uc *DefaultOrderUsecase) processWalletReleasesWithTracking(expiredOrders []dto.ExpiredOrderData) {
    if len(expiredOrders) == 0 {
        return
    }

    // Увеличиваем счетчик попыток для всех ордеров
    orderIDs := make([]string, len(expiredOrders))
    for i, order := range expiredOrders {
        orderIDs[i] = order.ID
    }
    if err := uc.OrderRepo.IncrementReleaseAttempts(context.Background(), orderIDs); err != nil {
        log.Printf("Failed to increment release attempts: %v", err)
    }

    walletReleases := make([]walletRequest.ReleaseRequest, len(expiredOrders))
    for i, order := range expiredOrders {
        walletReleases[i] = walletRequest.ReleaseRequest{
            TraderID:      order.TraderID,
            MerchantID:    order.MerchantID,
            OrderID:       order.ID,
            RewardPercent: order.TraderRewardPercent,
            PlatformFee:   order.PlatformFee,
        }
    }

    var successfulOrderIDs []string

    // Пытаемся батчевый release
    if err := uc.WalletHandler.BatchRelease(walletReleases); err != nil {
        log.Printf("Batch wallet release failed, falling back to individual releases: %v", err)
        
        // Fallback на индивидуальные запросы с детальным отслеживанием
        for _, release := range walletReleases {
            if err := uc.WalletHandler.Release(release); err != nil {
                log.Printf("CRITICAL: Failed to release wallet for order %s: %v", release.OrderID, err)
                // Отправляем в DLQ для ручной обработки
                uc.sendToDeadLetterQueue(release, err.Error())
            } else {
                successfulOrderIDs = append(successfulOrderIDs, release.OrderID)
                log.Printf("Successfully released wallet for order %s", release.OrderID)
            }
        }
    } else {
        // Все успешно разморозились
        for _, order := range expiredOrders {
            successfulOrderIDs = append(successfulOrderIDs, order.ID)
        }
        log.Printf("Batch wallet release completed successfully for %d orders", len(expiredOrders))
    }

    // Помечаем успешно разморженные ордера
    if len(successfulOrderIDs) > 0 {
        if err := uc.OrderRepo.MarkReleasedAt(context.Background(), successfulOrderIDs); err != nil {
            log.Printf("CRITICAL: Failed to mark orders as released: %v", err)
        } else {
            log.Printf("Marked %d orders as successfully released", len(successfulOrderIDs))
        }
    }

    // Логируем статистику
    failed := len(expiredOrders) - len(successfulOrderIDs)
    if failed > 0 {
        log.Printf("ALERT: %d/%d wallet releases FAILED", failed, len(expiredOrders))
    }
}

// sendToDeadLetterQueue - отправка проблемных ордеров в DLQ
func (uc *DefaultOrderUsecase) sendToDeadLetterQueue(release walletRequest.ReleaseRequest, errorMsg string) {
    dlqPayload := struct {
        OrderID   string    `json:"order_id"`
        Release   walletRequest.ReleaseRequest `json:"release"`
        Error     string    `json:"error"`
        Timestamp time.Time `json:"timestamp"`
    }{
        OrderID:   release.OrderID,
        Release:   release,
        Error:     errorMsg,
        Timestamp: time.Now(),
    }
    
    payload, _ := json.Marshal(dlqPayload)
    if err := uc.mqPub.Publish("orders.release.dlq", domain.Message{
        Key:   []byte(release.OrderID),
        Value: payload,
    }); err != nil {
        log.Printf("CRITICAL: Failed to send order %s to DLQ: %v", release.OrderID, err)
    }
}

// processWalletReleases - обработка разморозки средств
func (uc *DefaultOrderUsecase) processWalletReleases(expiredOrders []dto.ExpiredOrderData) {
    if len(expiredOrders) == 0 {
        return
    }

    walletReleases := make([]walletRequest.ReleaseRequest, len(expiredOrders))
    for i, order := range expiredOrders {
        walletReleases[i] = walletRequest.ReleaseRequest{
            TraderID:      order.TraderID,
            MerchantID:    order.MerchantID,
            OrderID:       order.ID,
            RewardPercent: order.TraderRewardPercent,
            PlatformFee:   order.PlatformFee,
        }
    }

    // Пытаемся батчевый release
    if err := uc.WalletHandler.BatchRelease(walletReleases); err != nil {
        log.Printf("Batch wallet release failed, falling back to individual releases: %v", err)
        
        // Fallback на индивидуальные запросы
        successCount := 0
        for _, release := range walletReleases {
            if err := uc.WalletHandler.Release(release); err != nil {
                log.Printf("Failed to release wallet for order %s: %v", release.OrderID, err)
            } else {
                successCount++
            }
        }
        
        log.Printf("Individual wallet releases completed: %d/%d successful", successCount, len(walletReleases))
    } else {
        log.Printf("Batch wallet release completed successfully for %d orders", len(walletReleases))
    }
}

// StartStuckOrdersMonitor - мониторинг зависших ордеров
func (uc *DefaultOrderUsecase) StartStuckOrdersMonitor(ctx context.Context) {
    ticker := time.NewTicker(2 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            stuckIDs, err := uc.OrderRepo.FindStuckOrders(ctx, 5) // максимум 5 попыток
            if err != nil {
                log.Printf("Failed to find stuck orders: %v", err)
                continue
            }
            
            if len(stuckIDs) > 0 {
                log.Printf("ALERT: Found %d stuck orders that need retry", len(stuckIDs))
                
                // Отправляем их на повторную обработку
                payload, _ := json.Marshal(stuckIDs)
                if err := uc.mqPub.Publish("orders.retry", domain.Message{
                    Key:   nil,
                    Value: payload,
                }); err != nil {
                    log.Printf("Failed to publish retry task: %v", err)
                }
            }
        }
    }
}

// StartRetryWorker - воркер для повторной обработки
func (uc *DefaultOrderUsecase) StartRetryWorker(ctx context.Context) {
    msgs, err := uc.mqSub.Subscribe("orders.retry", "order-retry-group")
    if err != nil {
        log.Fatalf("Failed to subscribe to retry topic: %v", err)
    }

    for {
        select {
        case <-ctx.Done():
            return
        case m, ok := <-msgs:
            if !ok {
                return
            }
            
            var orderIDs []string
            if err := json.Unmarshal(m.Value, &orderIDs); err != nil {
                log.Printf("Invalid retry payload: %v", err)
                continue
            }
            
            // Загружаем данные по ID и повторно обрабатываем
            expiredOrders, err := uc.OrderRepo.LoadExpiredOrderDataByIDs(ctx, orderIDs)
            if err != nil {
                log.Printf("Failed to load retry orders: %v", err)
                continue
            }
            
            log.Printf("Retrying %d stuck orders", len(expiredOrders))
            uc.processWalletReleasesWithTracking(expiredOrders)
        }
    }
}

// publishOrderEventsWithTracking - публикация событий с отслеживанием
func (uc *DefaultOrderUsecase) publishOrderEventsWithTracking(expiredOrders []dto.ExpiredOrderData) {
    if len(expiredOrders) == 0 {
        return
    }

    // Увеличиваем счетчик попыток публикации
    orderIDs := make([]string, len(expiredOrders))
    for i, order := range expiredOrders {
        orderIDs[i] = order.ID
    }
    if err := uc.OrderRepo.IncrementPublishAttempts(context.Background(), orderIDs); err != nil {
        log.Printf("Failed to increment publish attempts: %v", err)
    }

    // Формируем события
    events := make([]publisher.OrderEvent, len(expiredOrders))
    for i, order := range expiredOrders {
        events[i] = publisher.OrderEvent{
            OrderID:    order.ID,
            TraderID:   order.TraderID,
            Status:     "⛔️Отмена сделки",
            AmountFiat: order.AmountFiat,
            Currency:   order.Currency,
            BankName:   order.BankName,
            Phone:      order.Phone,
            CardNumber: order.CardNumber,
            Owner:      order.Owner,
        }
    }

    var successfulOrderIDs []string

    // Fallback на индивидуальные публикации с отслеживанием
    for _, event := range events {
		v, _ := json.Marshal(event)
        if err := uc.mqPub.Publish("order-events", domain.Message{Key: []byte(event.TraderID), Value: v}); err != nil {
            log.Printf("Failed to publish event for order %s: %v", event.OrderID, err)
        } else {
            successfulOrderIDs = append(successfulOrderIDs, event.OrderID)
            log.Printf("Successfully published event for order %s", event.OrderID)
        }
    }

    // Помечаем успешно опубликованные события
    if len(successfulOrderIDs) > 0 {
        if err := uc.OrderRepo.MarkPublishedAt(context.Background(), successfulOrderIDs); err != nil {
            log.Printf("Failed to mark orders as published: %v", err)
        } else {
            log.Printf("Marked %d orders as successfully published", len(successfulOrderIDs))
        }
    }

    // Логируем статистику
    failed := len(expiredOrders) - len(successfulOrderIDs)
    if failed > 0 {
        log.Printf("WARNING: %d/%d event publications FAILED", failed, len(expiredOrders))
    }
}

// sendCallbacksWithTracking - отправка callback'ов с отслеживанием
func (uc *DefaultOrderUsecase) sendCallbacksWithTracking(expiredOrders []dto.ExpiredOrderData) {
    // Фильтруем только те ордера, у которых есть callback URL
    var callbackOrders []dto.ExpiredOrderData
    var callbacks []CallbackRequest
    
    for _, order := range expiredOrders {
        if order.CallbackURL != "" {
            callbackOrders = append(callbackOrders, order)
            callbacks = append(callbacks, CallbackRequest{
                URL:             order.CallbackURL,
                MerchantOrderID: order.MerchantOrderID,
                OrderID:         order.ID, // Добавляем OrderID для отслеживания
                Status:          string(domain.StatusCanceled),
            })
        }
    }

    if len(callbacks) == 0 {
        log.Printf("No callbacks to send for expired orders")
        return
    }

    // Увеличиваем счетчик попыток для ордеров с callback'ами
    orderIDs := make([]string, len(callbackOrders))
    for i, order := range callbackOrders {
        orderIDs[i] = order.ID
    }
    if err := uc.OrderRepo.IncrementCallbackAttempts(context.Background(), orderIDs); err != nil {
        log.Printf("Failed to increment callback attempts: %v", err)
    }

    log.Printf("Sending %d callbacks with tracking", len(callbacks))

    // Отправляем callback'ы с детальным отслеживанием результатов
    results := uc.sendBatchCallbacksWithResults(callbacks)

    var successfulOrderIDs []string
    successCount := 0
    
    // Анализируем результаты
    for _, result := range results {
        if result.Success {
            successfulOrderIDs = append(successfulOrderIDs, result.OrderID)
            successCount++
            log.Printf("Successfully sent callback for order %s", result.OrderID)
        } else {
            log.Printf("Failed to send callback for order %s: %v", result.OrderID, result.Error)
        }
    }

    // Помечаем успешно отправленные callback'ы
    if len(successfulOrderIDs) > 0 {
        if err := uc.OrderRepo.MarkCallbacksSentAt(context.Background(), successfulOrderIDs); err != nil {
            log.Printf("Failed to mark callbacks as sent: %v", err)
        } else {
            log.Printf("Marked %d orders as callbacks sent", len(successfulOrderIDs))
        }
    }

    // Логируем статистику
    failed := len(callbacks) - successCount
    if failed > 0 {
        log.Printf("WARNING: %d/%d callbacks FAILED", failed, len(callbacks))
    }
}

// sendBatchCallbacksWithResults - отправка callback'ов с возвратом результатов
func (uc *DefaultOrderUsecase) sendBatchCallbacksWithResults(callbacks []CallbackRequest) []CallbackResult {
    results := make([]CallbackResult, len(callbacks))
    
    // Параллельная отправка callbacks с rate limiting
    semaphore := make(chan struct{}, 10) // Максимум 10 одновременных запросов
    var wg sync.WaitGroup
    var mu sync.Mutex

    for i, callback := range callbacks {
        wg.Add(1)
        go func(index int, cb CallbackRequest) {
            defer wg.Done()
            semaphore <- struct{}{} // Захватываем семафор
            defer func() { <-semaphore }() // Освобождаем семафор

            // Отправляем callback (retry уже внутри SendCallback)
            if err := notifier.SendCallback(cb.URL, cb.MerchantOrderID, cb.Status, 0, 0, 0); err != nil {
                mu.Lock()
                results[index] = CallbackResult{
                    OrderID:         cb.OrderID,
                    MerchantOrderID: cb.MerchantOrderID,
                    Success:         false, 
                    Error:           err.Error(),
                }
                mu.Unlock()
            } else {
                mu.Lock()
                results[index] = CallbackResult{
                    OrderID:         cb.OrderID,
                    MerchantOrderID: cb.MerchantOrderID,
                    Success:         true,
                }
                mu.Unlock()
            }
        }(i, callback)
    }

    wg.Wait()
    return results
}

// ApproveOrder - подтверждение заказа
func (uc *DefaultOrderUsecase) ApproveOrder(orderID string) error {
    order, err := uc.GetOrderByID(orderID)
    if err != nil {
        return err
    }

    if order.Order.Status != domain.StatusPending {
        return domain.ErrResolveDisputeFailed
    }

    // Подготавливаем commission users
    var commissionUsers []walletRequest.CommissionUser
    teamRelations, err := uc.TeamRelationsUsecase.GetRelationshipsByTraderID(order.BankDetail.TraderID)
    if err == nil {
        for _, teamRelation := range teamRelations {
            commissionUsers = append(commissionUsers, walletRequest.CommissionUser{
                UserID:     teamRelation.TeamLeadID,
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
                TraderID:        order.BankDetail.TraderID,
                MerchantID:      order.Order.MerchantInfo.MerchantID,
                OrderID:         order.Order.ID,
                RewardPercent:   order.Order.TraderReward,
                PlatformFee:     order.Order.PlatformFee,
                CommissionUsers: commissionUsers,
            },
        },
        EventData: &publisher.OrderEvent{
            OrderID:    order.Order.ID,
            TraderID:   order.BankDetail.TraderID,
            Status:     "✅Сделка закрыта",
            AmountFiat: order.Order.AmountInfo.AmountFiat,
            Currency:   order.Order.AmountInfo.Currency,
            BankName:   order.BankDetail.BankName,
            Phone:      order.BankDetail.Phone,
            CardNumber: order.BankDetail.CardNumber,
            Owner:      order.BankDetail.Owner,
        },
        CallbackData: &CallbackRequest{
            URL:             order.Order.CallbackUrl,
            MerchantOrderID: order.Order.MerchantInfo.MerchantOrderID,
            OrderID:         order.Order.ID,
            Status:          string(domain.StatusCompleted),
        },
    }

    return uc.ProcessOrderOperation(context.Background(), op)
}

// CancelOrder - отмена заказа
func (uc *DefaultOrderUsecase) CancelOrder(orderID string) error {
    order, err := uc.GetOrderByID(orderID)
    if err != nil {
        return err
    }

    if order.Order.Status != domain.StatusPending && order.Order.Status != domain.StatusDisputeCreated {
        return domain.ErrCancelOrder
    }

    op := &OrderOperation{
        OrderID:   orderID,
        Operation: "cancel",
        OldStatus: order.Order.Status,
        NewStatus: domain.StatusCanceled,
        WalletOp: &WalletOperation{
            Type: "release",
            Request: walletRequest.ReleaseRequest{
                TraderID:      order.BankDetail.TraderID,
                MerchantID:    order.Order.MerchantInfo.MerchantID,
                OrderID:       order.Order.ID,
                RewardPercent: 0, // При отмене не даем вознаграждение
                PlatformFee:   0, // При отмене не берем комиссию
            },
        },
        EventData: &publisher.OrderEvent{
            OrderID:    order.Order.ID,
            TraderID:   order.BankDetail.TraderID,
            Status:     "⛔️Отмена сделки",
            AmountFiat: order.Order.AmountInfo.AmountFiat,
            Currency:   order.Order.AmountInfo.Currency,
            BankName:   order.BankDetail.BankName,
            Phone:      order.BankDetail.Phone,
            CardNumber: order.BankDetail.CardNumber,
            Owner:      order.BankDetail.Owner,
        },
        CallbackData: &CallbackRequest{
            URL:             order.Order.CallbackUrl,
            MerchantOrderID: order.Order.MerchantInfo.MerchantOrderID,
            OrderID:         order.Order.ID,
            Status:          string(domain.StatusCanceled),
        },
    }

    return uc.ProcessOrderOperation(context.Background(), op)
}

func (uc *DefaultOrderUsecase) GetOrderStatistics(traderID string, dateFrom, dateTo time.Time) (*domain.OrderStatistics, error) {
	return uc.OrderRepo.GetOrderStatistics(traderID, dateFrom, dateTo)
}

func (uc *DefaultOrderUsecase) GetOrders(filter domain.Filter, sortField string, page, size int) ([]*domain.Order, int64, error) {
	return uc.OrderRepo.GetOrders(filter, sortField, page, size)
}

func (uc *DefaultOrderUsecase) GetAllOrders(input *orderdto.GetAllOrdersInput) (*orderdto.GetAllOrdersOutput, error) {
    // Валидация пагинации
    if input.Page < 1 {
        input.Page = 1
    }
    if input.Limit < 1 || input.Limit > 100 {
        input.Limit = 50 // дефолтное значение
    }

    // Преобразуем в фильтры репозитория
    filters := &domain.AllOrdersFilters{
        TraderID:         input.TraderID,
        MerchantID:       input.MerchantID,
        OrderID:          input.OrderID,
        MerchantOrderID:  input.MerchantOrderID,
        Status:           input.Status,
        BankCode:         input.BankCode,
        TimeOpeningStart: input.TimeOpeningStart,
        TimeOpeningEnd:   input.TimeOpeningEnd,
        AmountFiatMin:    input.AmountFiatMin,
        AmountFiatMax:    input.AmountFiatMax,
        Type:             input.Type,
        DeviceID:         input.DeviceID,
		PaymentSystem: 	  input.PaymentSystem,
    }

    // Вызываем репозиторий
    orders, total, err := uc.OrderRepo.GetAllOrders(filters, input.Sort, input.Page, input.Limit)
    if err != nil {
        return nil, err
    }

    // Рассчитываем данные пагинации
    totalPages := int32(total) / input.Limit
    if int32(total)%input.Limit > 0 {
        totalPages++
    }

    return &orderdto.GetAllOrdersOutput{
        Orders: orders,
        Pagination: orderdto.Pagination{
            CurrentPage:  input.Page,
            TotalPages:   totalPages,
            TotalItems:   int32(total),
            ItemsPerPage: input.Limit,
        },
    }, nil
}

////////////////////// Advanced Safe Order operations //////////////////////////

// OrderOperation - описание операции со сделкой
type OrderOperation struct {
    OrderID     string                    `json:"order_id"`
    Operation   string                    `json:"operation"` // "create", "approve", "cancel"
    OldStatus   domain.OrderStatus        `json:"old_status"`
    NewStatus   domain.OrderStatus        `json:"new_status"`
    WalletOp    *WalletOperation         `json:"wallet_op,omitempty"`
    EventData   *publisher.OrderEvent    `json:"event_data,omitempty"`
    CallbackData *CallbackRequest        `json:"callback_data,omitempty"`
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
    if err := uc.scheduleNonCriticalOperations(op); err != nil {
        log.Printf("Failed to schedule non-critical operations for order %s: %v", op.OrderID, err)
        // НЕ возвращаем ошибку - критичные операции уже выполнены
    }

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

/////////////////////////////// Асинхронная обработка некритичных операций //////////////////

// scheduleNonCriticalOperations - планирует некритичные операции
func (uc *DefaultOrderUsecase) scheduleNonCriticalOperations(op *OrderOperation) error {
    payload, _ := json.Marshal(op)
    return uc.mqPub.Publish("orders.processing", domain.Message{
        Key:   []byte(op.OrderID),
        Value: payload,
    })
}

// StartProcessingWorker - воркер для обработки некритичных операций
func (uc *DefaultOrderUsecase) StartProcessingWorker(ctx context.Context) {
    msgs, err := uc.mqSub.Subscribe("orders.processing", "order-processing-group")
    if err != nil {
        log.Fatalf("Failed to subscribe to processing topic: %v", err)
    }

    log.Println("Order processing worker started")

    for {
        select {
        case <-ctx.Done():
            return
        case m, ok := <-msgs:
            if !ok {
                return
            }

            var op OrderOperation
            if err := json.Unmarshal(m.Value, &op); err != nil {
                log.Printf("Invalid processing payload: %v", err)
                continue
            }

            uc.processNonCriticalOperations(&op)
        }
    }
}

// processNonCriticalOperations - обработка некритичных операций
// processNonCriticalOperations - обработка некритичных операций
func (uc *DefaultOrderUsecase) processNonCriticalOperations(op *OrderOperation) {
    state, err := uc.getTransactionState(op.OrderID)
    if err != nil {
        log.Printf("Failed to get transaction state for %s: %v", op.OrderID, err)
        return
    }

    var updated bool

    // Публикация в Kafka
    if !state.EventPublished && op.EventData != nil {
        if err := uc.publishOrderEvent(op.EventData); err != nil {
            log.Printf("Failed to publish event for order %s: %v", op.OrderID, err)
        } else {
            if err := uc.markEventPublished(op.OrderID); err != nil {
                log.Printf("Failed to mark event as published for order %s: %v", op.OrderID, err)
            } else {
                log.Printf("Published event for order %s", op.OrderID)
                updated = true
            }
        }
    }

    // Отправка callback
    if !state.CallbackSent && op.CallbackData != nil && op.CallbackData.URL != "" {
        if err := notifier.SendCallback(op.CallbackData.URL, op.CallbackData.MerchantOrderID, op.CallbackData.Status, 0, 0, 0); err != nil {
            log.Printf("Failed to send callback for order %s: %v", op.OrderID, err)
        } else {
            if err := uc.markCallbackSent(op.OrderID); err != nil {
                log.Printf("Failed to mark callback as sent for order %s: %v", op.OrderID, err)
            } else {
                log.Printf("Sent callback for order %s", op.OrderID)
                updated = true
            }
        }
    }

    // Если все операции завершены, отмечаем транзакцию как завершенную
    if updated {
        // Проверяем, завершены ли все некритичные операции
        if err := uc.checkAndMarkCompleted(op.OrderID); err != nil {
            log.Printf("Failed to check completion status for order %s: %v", op.OrderID, err)
        }
    }
}

// checkAndMarkCompleted - проверка и отметка завершения всех операций
func (uc *DefaultOrderUsecase) checkAndMarkCompleted(orderID string) error {
    state, err := uc.getTransactionState(orderID)
    if err != nil {
        return err
    }

    // Проверяем, завершены ли все нужные операции
    allCompleted := state.StatusChanged && state.WalletProcessed
    
    // Если есть события для публикации, они должны быть опубликованы
    if state.EventPublished {
        allCompleted = allCompleted && state.EventPublished
    }
    
    // Если есть callback для отправки, они должны быть отправлены
    if state.CallbackSent {
        allCompleted = allCompleted && state.CallbackSent
    }

    // Если все завершено и еще не отмечено как завершенное
    if allCompleted && state.CompletedAt == nil {
        if err := uc.OrderRepo.MarkCompleted(orderID); err != nil {
            return err
        }
        log.Printf("Marked order %s as fully completed", orderID)
    }

    return nil
}

// getTransactionState - получение состояния транзакции
func (uc *DefaultOrderUsecase) getTransactionState(orderID string) (*domain.OrderTransactionStateModel, error) {
    return uc.OrderRepo.GetTransactionState(orderID)
}

// publishOrderEvent - публикация события заказа в Kafka
func (uc *DefaultOrderUsecase) publishOrderEvent(event *publisher.OrderEvent) error {
    eventJSON, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    return uc.mqPub.Publish("order-events", domain.Message{
        Key:   []byte(event.TraderID),
        Value: eventJSON,
    })
}

// markEventPublished - отметка успешной публикации события
func (uc *DefaultOrderUsecase) markEventPublished(orderID string) error {
    return uc.OrderRepo.MarkEventPublished(orderID)
}

// markCallbackSent - отметка успешной отправки callback
func (uc *DefaultOrderUsecase) markCallbackSent(orderID string) error {
    return uc.OrderRepo.MarkCallbackSent(orderID)
}

//////////////////////////// Мониторинг несоответствий ///////////////////

// StartConsistencyMonitor - мониторинг консистентности статусов и кошельков
func (uc *DefaultOrderUsecase) StartConsistencyMonitor(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := uc.checkOrderWalletConsistency(); err != nil {
                log.Printf("Consistency check failed: %v", err)
            }
        }
    }
}

// checkOrderWalletConsistency - проверка соответствия статусов и кошельков
func (uc *DefaultOrderUsecase) checkOrderWalletConsistency() error {
    // Проверяем несоответствия между статусами ордеров и состоянием кошельков
    inconsistent, err := uc.OrderRepo.FindInconsistentOrders()
    if err != nil {
        return err
    }

    if len(inconsistent) > 0 {
        log.Printf("ALERT: Found %d inconsistent orders", len(inconsistent))
        
        // Отправляем на исправление
        payload, _ := json.Marshal(inconsistent)
        if err := uc.mqPub.Publish("orders.fix-consistency", domain.Message{
            Value: payload,
        }); err != nil {
            log.Printf("Failed to publish consistency fix task: %v", err)
        }
    }

    return nil
}