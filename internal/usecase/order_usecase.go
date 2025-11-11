package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"math"
	"math/rand"
	"time"

	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	bankdetaildto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/bank_detail"
	orderdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/order"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderUsecase interface {
	CreateOrder(input *orderdto.CreateOrderInput) (*orderdto.OrderOutput, error)
	GetOrderByID(orderID string) (*domain.Order, error)
	GetOrderByMerchantOrderID(merchantOrderID string) (*domain.Order, error)
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
	ProcessAutomaticPayment(ctx context.Context, req *AutomaticPaymentRequest) (*domain.AutomaticPaymentResult, error)
}

type DefaultOrderUsecase struct {
	OrderRepo 			domain.OrderRepository
	WalletHandler   	*handlers.HTTPWalletHandler
	TrafficUsecase  	TrafficUsecase
	BankDetailUsecase 	BankDetailUsecase
	TeamRelationsUsecase TeamRelationsUsecase
	Publisher 			*publisher.KafkaPublisher
}

func NewDefaultOrderUsecase(
	orderRepo domain.OrderRepository, 
	walletHandler *handlers.HTTPWalletHandler,
	trafficUsecase TrafficUsecase,
	bankDetailUsecase BankDetailUsecase,
	kafkaPublisher *publisher.KafkaPublisher,
	teamRelationsUsecase TeamRelationsUsecase) *DefaultOrderUsecase {

	return &DefaultOrderUsecase{
		OrderRepo: orderRepo,
		WalletHandler: walletHandler,
		TrafficUsecase: trafficUsecase,
		BankDetailUsecase: bankDetailUsecase,
		Publisher: kafkaPublisher,
		TeamRelationsUsecase: teamRelationsUsecase,
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


func (uc *DefaultOrderUsecase) GetOrderByID(orderID string) (*domain.Order, error) {
	return uc.OrderRepo.GetOrderByID(orderID)
}

func (uc *DefaultOrderUsecase) GetOrderByMerchantOrderID(merchantOrderID string) (*domain.Order, error) {
	return uc.OrderRepo.GetOrderByMerchantOrderID(merchantOrderID)
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

func (uc *DefaultOrderUsecase) ApproveOrder(orderID string) error {
	// Find exact order
	order, err := uc.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != domain.StatusPending {
		return domain.ErrResolveDisputeFailed
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
	op := &OrderOperation{
        OrderID:   orderID,
        Operation: "approve",
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
		CreatedAt: time.Now(),
    }

	if err := uc.ProcessOrderOperation(context.Background(), op); err != nil {
		return err
	}

	go func(event publisher.OrderEvent){
		if err := uc.Publisher.PublishOrder(event); err != nil {
			slog.Error("failed to publish kafka OrderEvent", "stage", "approving", "error", err.Error())
		}
	}(publisher.OrderEvent{
		OrderID: order.ID,
		TraderID: order.RequisiteDetails.TraderID,
		Status: "✅Сделка закрыта",
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
			string(domain.StatusCompleted),
			0, 0, 0,
		)
	}

	return nil
}

func (uc *DefaultOrderUsecase) CancelOrder(orderID string) error {
	// Find exact order
	order, err := uc.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != domain.StatusPending && order.Status != domain.StatusDisputeCreated{
		return domain.ErrCancelOrder
	}

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

	return nil
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