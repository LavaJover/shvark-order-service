package usecase

import (
	"context"
	"fmt"
	"log"
	"log/slog"
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
	Publisher 			*kafka.KafkaPublisher
}

func NewDefaultOrderUsecase(
	orderRepo domain.OrderRepository, 
	walletHandler *handlers.HTTPWalletHandler,
	trafficUsecase domain.TrafficUsecase,
	bankDetailUsecase BankDetailUsecase,
	kafkaPublisher *kafka.KafkaPublisher,
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

func (uc *DefaultOrderUsecase) FilterByMaxOrdersSimulateosly(bankDetails []*domain.BankDetail) ([]*domain.BankDetail, error) {
	result := make([]*domain.BankDetail, 0)
	for _, bankDetail := range bankDetails {
		orders, err := uc.OrderRepo.GetOrdersByBankDetailID(bankDetail.ID)
		if err != nil {
			return nil, err
		}
		ordersCreated := make([]*domain.Order, 0)
		for _, order := range orders {
			if order.Status == domain.StatusPending {
				ordersCreated = append(ordersCreated, order)
			}
		}
		fmt.Printf("Orders created: %d. Orders max simultaneosly: %d\n", len(ordersCreated), bankDetail.MaxOrdersSimultaneosly)
		if len(ordersCreated) < int(bankDetail.MaxOrdersSimultaneosly) {
			result = append(result, bankDetail)
		}
	}

	return result, nil
}

func (uc *DefaultOrderUsecase) FilterByMaxAmountDay(bankDetails []*domain.BankDetail, amountFiat float64) ([]*domain.BankDetail, error) {
	result := make([]*domain.BankDetail, 0)
	for _, bankDetail := range bankDetails {
		orders, err := uc.OrderRepo.GetOrdersByBankDetailID(bankDetail.ID)
		if err != nil {
			return nil, err
		}
		ordersSucceedSummary := float64(0.)
		now := time.Now()
		startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		for _, order := range orders {
			if (order.Status == domain.StatusCompleted || order.Status == domain.StatusPending) && (order.UpdatedAt.After(startOfToday)) {
				ordersSucceedSummary += order.AmountInfo.AmountFiat
			}
		}
		fmt.Printf("Max amount a day: %f. Current summary amount: %f\n", bankDetail.MaxAmountDay, ordersSucceedSummary)
		if ordersSucceedSummary + amountFiat <= float64(bankDetail.MaxAmountDay) {
			result = append(result, bankDetail)
		}
	}

	return result, nil
}

func (uc *DefaultOrderUsecase) FilterByMaxAmountMonth(bankDetails []*domain.BankDetail, amountFiat float64) ([]*domain.BankDetail, error) {
	result := make([]*domain.BankDetail, 0)
	for _, bankDetail := range bankDetails {
		orders, err := uc.OrderRepo.GetOrdersByBankDetailID(bankDetail.ID)
		if err != nil {
			return nil, err
		}
		ordersSucceedSummary := float64(0.)
		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		for _, order := range orders {
			if (order.Status == domain.StatusCompleted || order.Status == domain.StatusPending) && (order.UpdatedAt.After(startOfMonth)) {
				ordersSucceedSummary += order.AmountInfo.AmountFiat
			}
		}
		fmt.Printf("Max amount a month: %f. Current summary month: %f\n", bankDetail.MaxAmountMonth, ordersSucceedSummary)
		if ordersSucceedSummary + amountFiat <= float64(bankDetail.MaxAmountMonth) {
			result = append(result, bankDetail)
		}
	}

	return result, nil	
}

func (uc *DefaultOrderUsecase) FilterByDelay(bankDetails []*domain.BankDetail) ([]*domain.BankDetail, error) {
	result := make([]*domain.BankDetail, 0)
	for _, bankDetail := range bankDetails {
		var latestOrder *domain.Order
		orders, err := uc.OrderRepo.GetOrdersByBankDetailID(bankDetail.ID)
		if err != nil {
			return nil, err
		}
		for _, order := range orders {
			if order.Status != domain.StatusCompleted {
				continue
			}

			if latestOrder == nil || order.UpdatedAt.After(latestOrder.UpdatedAt) {
				latestOrder = order
			}
		}
		if latestOrder == nil || time.Since(latestOrder.UpdatedAt)>=bankDetail.Delay{
			result = append(result, bankDetail)
		}
	}
	return result, nil
}

func (uc *DefaultOrderUsecase) FilterByMaxQuantityDay(bankDetails []*domain.BankDetail) ([]*domain.BankDetail, error) {
	result := make([]*domain.BankDetail, 0)
	for _, bankDetail := range bankDetails {
		orders, err := uc.OrderRepo.GetOrdersByBankDetailID(bankDetail.ID)
		if err != nil {
			return nil, err
		}
		ordersQuantityDay := 0 
		for _, order := range orders {
			if order.Status == domain.StatusCanceled {
				continue
			}
			now := time.Now()
			startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			if (order.Status == domain.StatusPending || order.Status == domain.StatusCompleted) && (order.UpdatedAt.After(startOfToday)) {
				ordersQuantityDay++
				continue
			}
		}
		fmt.Printf("Max quantity a day: %d. Current daily quantity: %d\n", bankDetail.MaxQuantityDay, ordersQuantityDay)
		if ordersQuantityDay + 1 <= int(bankDetail.MaxQuantityDay) {
			result = append(result, bankDetail)
		}
	}

	return result, nil
}

func (uc *DefaultOrderUsecase) FilterByMaxQuantityMonth(bankDetails []*domain.BankDetail) ([]*domain.BankDetail, error) {
	result := make([]*domain.BankDetail, 0)
	for _, bankDetail := range bankDetails {
		orders, err := uc.OrderRepo.GetOrdersByBankDetailID(bankDetail.ID)
		if err != nil {
			return nil, err
		}
		ordersQuantityMonth := 0 
		for _, order := range orders {
			if order.Status == domain.StatusCanceled {
				continue
			}
			now := time.Now()
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			if (order.Status == domain.StatusPending || order.Status == domain.StatusCompleted) && (order.UpdatedAt.After(startOfMonth)) {
				ordersQuantityMonth++
				continue
			}
		}
		fmt.Printf("Max quantity a month: %d. Current monthly quantity: %d\n", bankDetail.MaxQuantityMonth, ordersQuantityMonth)
		if ordersQuantityMonth + 1 <= int(bankDetail.MaxQuantityMonth) {
			result = append(result, bankDetail)
		}
	}

	return result, nil	
}

func (uc *DefaultOrderUsecase)FilterByTraderBalance(bankDetails []*domain.BankDetail, amountCrypto float64) ([]*domain.BankDetail, error) {
	result := make([]*domain.BankDetail, 0)
	for _, bankDetail := range bankDetails {
		traderBalance, err := uc.WalletHandler.GetTraderBalance(bankDetail.TraderID)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Trader %s balance: %f\n. Order: %f\n", bankDetail.TraderID, traderBalance, amountCrypto)
		if traderBalance >= amountCrypto {
			result = append(result, bankDetail)
		}
	}
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
	bankDetails, err = uc.FilterByTraderBalance(bankDetails, input.AmountCrypto)
	if err != nil {
		return nil, err
	}
	if len(bankDetails) == 0 {
		log.Printf("Отсеились по балансу трейдеров\n")
	}

	// 2) Filter by MaxOrdersSimultaneosly
	bankDetails, err = uc.FilterByMaxOrdersSimulateosly(bankDetails)
	if err != nil {
		return nil, err
	}
	if len(bankDetails) == 0 {
		log.Printf("Отсеились по одновременным сделкам\n")
	}
	// 3) Filter by MaxAmountDay
	bankDetails, err = uc.FilterByMaxAmountDay(bankDetails, input.AmountFiat)
	if err != nil {
		return nil, err
	}
	if len(bankDetails) == 0 {
		log.Printf("Отсеились по сумме в день\n")
	}
	// 4) Filter by MaxAmountMonth
	bankDetails, err = uc.FilterByMaxAmountMonth(bankDetails, input.AmountFiat)
	if err != nil {
		return nil, err
	}
	if len(bankDetails) == 0 {
		log.Printf("Отсеились по сумме в месяц\n")
	}
	// 5) Filter by delay
	bankDetails, err = uc.FilterByDelay(bankDetails)
	if err != nil {
		return nil, err
	}
	if len(bankDetails) == 0 {
		log.Printf("Отсеились по задержке\n")
	}
	// 6) Filter by MaxQuantityDay
	bankDetails, err = uc.FilterByMaxQuantityDay(bankDetails)
	if err != nil {
		return nil, err
	}
	if len(bankDetails) == 0 {
		log.Printf("Отсеились по количеству в день\n")
	}

	// 7) Filter by MaxQuantityMonth
	bankDetails, err = uc.FilterByMaxQuantityMonth(bankDetails)
	if err != nil {
		return nil, err
	}
	if len(bankDetails) == 0 {
		log.Printf("Отсеились по количеству в месяц\n")
	}

	// 8) Filter by active order with equal amount fiat
	tempBankDetails, err := uc.FilterByEqualAmountFiat(bankDetails, input.AmountFiat)
	if err != nil {
		return nil, err
	}
	// Если shuffle не задан, то пропускаем сериб проверок с рекалькуляцией
	for addFiat := range input.AdvancedParams.Shuffle {
		tempBankDetails, err = uc.FilterByEqualAmountFiat(bankDetails, input.AmountFiat + float64(addFiat))
		if err != nil {
			return nil, err
		}
		if len(tempBankDetails) != 0 {
			input.AmountFiat += float64(addFiat)
			if addFiat != 0 {
				input.AdvancedParams.Recalculated = true
			}else {
				input.AdvancedParams.Recalculated = false
			}
			break
		}
	}
	bankDetails = tempBankDetails
	if len(bankDetails) == 0 {
		return nil, domain.ErrNoAvailableBankDetails
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
	// check idempotency by client_id
	if createOrderInput.ClientID != "" {
		if err := uc.CheckIdempotency(createOrderInput.ClientID); err != nil {
			return nil, err
		}
	}

	// searching for eligible bank details due to order query parameters
	bankDetails, err := uc.FindEligibleBankDetails(createOrderInput)
	if err != nil {
		return nil, status.Error(codes.NotFound, "no eligible bank detail" + err.Error())
	}
	if len(bankDetails) != 0 {
		log.Printf("Реквизиты для заявки не найдены!\n")
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
	chosenBankDetail, err := uc.PickBestBankDetail(bankDetails, createOrderInput.MerchantID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to pick best bank detail for order")
	}

	// Get trader reward percent and save to order
	traffic, err := uc.TrafficUsecase.GetTrafficByTraderMerchant(chosenBankDetail.TraderID, createOrderInput.MerchantID)
	if err != nil {
		return nil, err
	}
	traderReward := traffic.TraderRewardPercent
	platformFee := traffic.PlatformFee

	order := domain.Order{
		ID: uuid.New().String(),
		Status: domain.StatusPending,
		MerchantInfo: domain.MerchantInfo{
			MerchantID: createOrderInput.MerchantID,
			MerchantOrderID: createOrderInput.MerchantOrderID,
			ClientID: createOrderInput.ClientID,
		},
		AmountInfo: domain.AmountInfo{
			AmountFiat: createOrderInput.AmountFiat,
			AmountCrypto: createOrderInput.AmountCrypto,
			CryptoRate: createOrderInput.CryptoRate,
			Currency: createOrderInput.Currency,
		},
		BankDetailID: chosenBankDetail.ID,
		Type: createOrderInput.Type,
		Recalculated: createOrderInput.Recalculated,
		Shuffle: createOrderInput.Shuffle,
		TraderReward: traderReward,
		PlatformFee: platformFee,
		CallbackUrl: createOrderInput.CallbackUrl,
		ExpiresAt: createOrderInput.ExpiresAt,
	}
	err = uc.OrderRepo.CreateOrder(&order)
	if err != nil {
		return nil, err
	}

	// Freeze crypto
	if err := uc.WalletHandler.Freeze(chosenBankDetail.TraderID, order.ID, createOrderInput.AmountCrypto); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = uc.Publisher.Publish(kafka.OrderEvent{
		OrderID: order.ID,
		TraderID: chosenBankDetail.TraderID,
		Status: "🔥Новая сделка",
		AmountFiat: order.AmountInfo.AmountFiat,
		Currency: order.AmountInfo.Currency,
		BankName: chosenBankDetail.BankName,
		Phone: chosenBankDetail.Phone,
		CardNumber: chosenBankDetail.CardNumber,
		Owner: chosenBankDetail.Owner,
	}); err != nil {
		slog.Error("failed to publish event", "error", err.Error())
	}

	if order.CallbackUrl != "" {
		notifier.SendCallback(
			order.CallbackUrl,
			order.MerchantInfo.MerchantOrderID,
			string(domain.StatusPending),
			0, 0, 0,
		)
	}

	return &orderdto.OrderOutput{
		Order: order,
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

	if order.Order.Status != domain.StatusPending {
		return domain.ErrResolveDisputeFailed
	}

	// Search for team relations to find commission users
	var commissionUsers []walletRequest.CommissionUser
	teamRelations, err := uc.TeamRelationsUsecase.GetRelationshipsByTraderID(order.BankDetail.TraderID)
	if err == nil {
		for _, teamRelation := range teamRelations {
			commissionUsers = append(commissionUsers, walletRequest.CommissionUser{
				UserID: teamRelation.TeamLeadID,
				Commission: teamRelation.TeamRelationshipRapams.Commission,
			})
		}
	}
	// make request to wallet-service to release order
	releaseRequest := walletRequest.ReleaseRequest{
		TraderID: order.BankDetail.TraderID,
		MerchantID: order.Order.MerchantInfo.MerchantID,
		OrderID: order.Order.ID,
		RewardPercent: order.Order.TraderReward,
		PlatformFee: order.Order.PlatformFee,
		CommissionUsers: commissionUsers,
	}
	if err := uc.WalletHandler.Release(releaseRequest); err != nil {
		return err
	}

	// Set order status to SUCCEED
	if err := uc.OrderRepo.UpdateOrderStatus(orderID, domain.StatusCompleted); err != nil {
		return err
	}

	if err = uc.Publisher.Publish(kafka.OrderEvent{
		OrderID: order.Order.ID,
		TraderID: order.BankDetail.TraderID,
		Status: "✅Сделка закрыта",
		AmountFiat: order.Order.AmountInfo.AmountFiat,
		Currency: order.Order.AmountInfo.Currency,
		BankName: order.BankDetail.BankName,
		Phone: order.BankDetail.Phone,
		CardNumber: order.BankDetail.CardNumber,
		Owner: order.BankDetail.Owner,
	}); err != nil {
		slog.Error("failed to publish event", "error", err.Error())
	}

	if order.Order.CallbackUrl != "" {
		notifier.SendCallback(
			order.Order.CallbackUrl,
			order.Order.MerchantInfo.MerchantOrderID,
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

	if order.Order.Status != domain.StatusPending && order.Order.Status != domain.StatusDisputeCreated{
		return domain.ErrCancelOrder
	}

	// Set order status to CANCELED
	if err := uc.OrderRepo.UpdateOrderStatus(orderID, domain.StatusCanceled); err != nil {
		return err
	}
	// Search for team relations to find commission users
	releaseRequest := walletRequest.ReleaseRequest{
		TraderID: order.BankDetail.TraderID,
		MerchantID: order.Order.MerchantInfo.MerchantID,
		OrderID: order.Order.ID,
		RewardPercent: 1,
		PlatformFee: 1,
	}
	if err := uc.WalletHandler.Release(releaseRequest); err != nil {
		return err
	}

	if err = uc.Publisher.Publish(kafka.OrderEvent{
		OrderID: order.Order.ID,
		TraderID: order.BankDetail.TraderID,
		Status: "⛔️Отмена сделки",
		AmountFiat: order.Order.AmountInfo.AmountFiat,
		Currency: order.Order.AmountInfo.Currency,
		BankName: order.BankDetail.BankName,
		Phone: order.BankDetail.Phone,
		CardNumber: order.BankDetail.CardNumber,
		Owner: order.BankDetail.Owner,
	}); err != nil {
		slog.Error("failed to publish event", "error", err.Error())
	}

	if order.Order.CallbackUrl != "" {
		notifier.SendCallback(
			order.Order.CallbackUrl,
			order.Order.MerchantInfo.MerchantOrderID,
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