package usecase

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	publisher "github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	bankdetaildto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/bank_detail"
	orderdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/order"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		if traffic.ActivityParams.AntifraudUnlocked && traffic.ActivityParams.ManuallyUnlocked && traffic.ActivityParams.MerchantUnlocked && traffic.ActivityParams.TraderUnlocked {
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

func (uc *DefaultOrderUsecase) FindEligibleBankDetails(input *orderdto.CreatePayInOrderInput) ([]*domain.BankDetail, error) {
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

func (uc *DefaultOrderUsecase) CreatePayInOrder(createOrderInput *orderdto.CreatePayInOrderInput) (*orderdto.OrderOutput, error) {
    start := time.Now()
    slog.Info("CreateOrder started")
    
    // check idempotency 
    if createOrderInput.ClientID != "" {
        t := time.Now()
        if err := uc.CheckIdempotency(createOrderInput.ClientID); err != nil {
            return nil, err
        }
        slog.Info("CheckIdempotency done", "elapsed", time.Since(t))
    }

    // searching for eligible bank details
    t := time.Now()
    bankDetails, err := uc.FindEligibleBankDetailsWithLock(createOrderInput)
    if err != nil {
        return nil, status.Error(codes.NotFound, "no eligible bank detail"+err.Error())
    }
    slog.Info("FindEligibleBankDetailsWithLock done", "elapsed", time.Since(t))
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
        Type:          domain.TypePayIn,
        Recalculated:  createOrderInput.Recalculated,
        Shuffle:       createOrderInput.Shuffle,
        TraderReward:  traderReward,
        PlatformFee:   platformFee,
        CallbackUrl:   createOrderInput.CallbackUrl,
        ExpiresAt:     time.Now().Add(traffic.BusinessParams.MerchantDealsDuration),

        RequisiteDetails: domain.RequisiteDetails{
            TraderID: chosenBankDetail.TraderID,
            CardNumber: chosenBankDetail.CardNumber,
            Phone: chosenBankDetail.Phone,
            Owner: chosenBankDetail.Owner,
            PaymentSystem: chosenBankDetail.PaymentSystem,
            BankName: chosenBankDetail.BankName,
            BankCode: chosenBankDetail.BankCode,
            NspkCode: chosenBankDetail.NspkCode,
            DeviceID: chosenBankDetail.DeviceID,
        },
        Metrics: domain.Metrics{},
    }
    
    t = time.Now()
    err = uc.OrderRepo.CreateOrder(&order)
    if err != nil {
        return nil, err
    }
    slog.Info("OrderRepo.CreateOrder done", "elapsed", time.Since(t))

    // Freeze crypto
    t = time.Now()
    if err := uc.WalletHandler.Freeze(chosenBankDetail.TraderID, order.ID, createOrderInput.AmountCrypto); err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    slog.Info("WalletHandler.Freeze done", "elapsed", time.Since(t))

    // Publish to kafka асинхронно
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

    slog.Info("CreateOrder finished", "total_elapsed", time.Since(start))

    return &orderdto.OrderOutput{
        Order:     order,
        BankDetail: *chosenBankDetail,
    }, nil
}

/////////////////// ATOMIC ///////////////////////////
// CreateOrderAtomic атомарно создает заказ в транзакции
func (uc *DefaultOrderUsecase) CreatePayInOrderAtomic(createOrderInput *orderdto.CreatePayInOrderInput) (*orderdto.OrderOutput, error) {
    start := time.Now()
    slog.Info("CreateOrderAtomic started")
    
    // Начинаем транзакцию
    txRepo, err := uc.OrderRepo.BeginTx()
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    
    // Гарантируем откат в случае ошибки
    var committed bool
    defer func() {
        if !committed {
            if rollbackErr := txRepo.Rollback(); rollbackErr != nil {
                slog.Error("Failed to rollback transaction", "error", rollbackErr)
            }
        }
    }()

    // check idempotency в транзакции
    if createOrderInput.ClientID != "" {
        if err := uc.checkIdempotencyInTx(txRepo, createOrderInput.ClientID); err != nil {
            return nil, err
        }
    }

    // Создаем BankDetailRepo с транзакцией
    bankDetailRepo := uc.BankDetailUsecase.(*usecase.DefaultBankDetailUsecase).GetBankDetailRepo()
    bankDetailRepoWithTx := bankDetailRepo.WithTx(txRepo)

    // Поиск реквизитов в транзакции с блокировкой
    bankDetails, err := uc.findEligibleBankDetailsInTx(bankDetailRepoWithTx, createOrderInput)
    if err != nil {
        return nil, status.Error(codes.NotFound, "no eligible bank detail"+err.Error())
    }
    
    if len(bankDetails) == 0 {
        log.Printf("Реквизиты для заявки не найдены!\n")
        return nil, fmt.Errorf("no available bank details")
    }
    log.Printf("Для заявки найдены доступные реквизиты!\n")

    // Выбор лучшего реквизита
    chosenBankDetail, err := uc.PickBestBankDetail(bankDetails, createOrderInput.MerchantID)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "failed to pick best bank detail for order")
    }

    // Получение трафика
    traffic, err := uc.TrafficUsecase.GetTrafficByTraderMerchant(chosenBankDetail.TraderID, createOrderInput.MerchantID)
    if err != nil {
        return nil, err
    }
    
    traderReward := traffic.TraderRewardPercent
    platformFee := traffic.PlatformFee

    // Создаем заказ
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
        Type:          domain.TypePayIn,
        Recalculated:  createOrderInput.Recalculated,
        Shuffle:       createOrderInput.Shuffle,
        TraderReward:  traderReward,
        PlatformFee:   platformFee,
        CallbackUrl:   createOrderInput.CallbackUrl,
        ExpiresAt:     time.Now().Add(traffic.BusinessParams.MerchantDealsDuration),

        RequisiteDetails: domain.RequisiteDetails{
            TraderID: chosenBankDetail.TraderID,
            CardNumber: chosenBankDetail.CardNumber,
            Phone: chosenBankDetail.Phone,
            Owner: chosenBankDetail.Owner,
            PaymentSystem: chosenBankDetail.PaymentSystem,
            BankName: chosenBankDetail.BankName,
            BankCode: chosenBankDetail.BankCode,
            NspkCode: chosenBankDetail.NspkCode,
            DeviceID: chosenBankDetail.DeviceID,
        },
        Metrics: domain.Metrics{},
    }

    // Сохраняем заказ в транзакции
    err = txRepo.CreateOrderInTx(&order)
    if err != nil {
        return nil, err
    }

    // Коммитим транзакцию
    if err := txRepo.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }
    committed = true

    // Отправляем колбэк о создании
    if createOrderInput.AdvancedParams.CallbackUrl != "" {
        notifier.SendCallback(
            createOrderInput.AdvancedParams.CallbackUrl,
            createOrderInput.MerchantOrderID,
            string(domain.StatusCreated),
            0, 0, 0,
        )
    }

    // Freeze crypto (после коммита транзакции)
    if err := uc.WalletHandler.Freeze(chosenBankDetail.TraderID, order.ID, createOrderInput.AmountCrypto); err != nil {
        // Если freeze не удался, отменяем заказ
        uc.cancelOrderDueToFreezeFailure(&order, err)
        return nil, status.Error(codes.Internal, err.Error())
    }

    // Публикация в Kafka и колбэки (асинхронно)
    uc.sendOrderNotifications(&order, chosenBankDetail)

    slog.Info("CreateOrderAtomic finished", "total_elapsed", time.Since(start))

    return &orderdto.OrderOutput{
        Order:     order,
        BankDetail: *chosenBankDetail,
    }, nil
}
// Вспомогательные методы для атомарного создания
func (uc *DefaultOrderUsecase) findEligibleBankDetailsInTx(bankDetailRepo domain.BankDetailRepository, input *orderdto.CreatePayInOrderInput) ([]*domain.BankDetail, error) {
    bankDetails, err := bankDetailRepo.FindSuitableBankDetailsInTx(
        &domain.SuitablleBankDetailsQuery{
            AmountFiat:    input.AmountFiat,
            Currency:      input.Currency,
            PaymentSystem: input.PaymentSystem,
            BankCode:      input.BankInfo.BankCode,
            NspkCode:      input.BankInfo.NspkCode,
        },
    )
    if err != nil {
        return nil, err
    }

    if len(bankDetails) == 0 {
        log.Printf("Отсеились по статическим параметрам\n")
        return []*domain.BankDetail{}, nil
    }

    // Filter by Traffic
    bankDetails, err = uc.FilterByTraffic(bankDetails, input.MerchantParams.MerchantID)
    if err != nil {
        return nil, err
    }
    if len(bankDetails) == 0 {
        log.Printf("Отсеились по трафику\n")
    }

    // Filter by Trader Available balances
    bankDetails, err = uc.FilterByTraderBalanceOptimal(bankDetails, input.AmountCrypto)
    if err != nil {
        return nil, err
    }
    if len(bankDetails) == 0 {
        log.Printf("Отсеились по балансу трейдеров\n")
    }

    return bankDetails, nil
}

func (uc *DefaultOrderUsecase) checkIdempotencyInTx(orderRepo domain.OrderRepository, clientID string) error {
    orders, err := orderRepo.GetCreatedOrdersByClientIDInTx(clientID)
    if len(orders) != 0 || err != nil {
        return status.Errorf(codes.FailedPrecondition, "payment order already exists for client: %s", clientID)
    }
    return nil
}

func (uc *DefaultOrderUsecase) FindEligibleBankDetailsWithLock(input *orderdto.CreatePayInOrderInput) ([]*domain.BankDetail, error) {
    // Используем метод с блокировкой вместо обычного
    bankDetails, err := uc.BankDetailUsecase.FindSuitableBankDetailsWithLock(
        &bankdetaildto.FindSuitableBankDetailsInput{
            AmountFiat:    input.AmountFiat,
            Currency:      input.Currency,
            PaymentSystem: input.PaymentSystem,
            BankCode:      input.BankInfo.BankCode,
            NspkCode:      input.BankInfo.NspkCode,
        },
    )
    if err != nil {
        return nil, err
    }

    if len(bankDetails) == 0 {
        log.Printf("Отсеились по статическим параметрам\n")
        return []*domain.BankDetail{}, nil
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