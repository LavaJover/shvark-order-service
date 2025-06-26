package usecase

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/client"
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/notifier"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/usdt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DefaultOrderUsecase struct {
	OrderRepo 		domain.OrderRepository
	BankDetailRepo 	domain.BankDetailRepository
	BankingClient   *client.BankingClient
	WalletHandler   *handlers.HTTPWalletHandler
	TrafficUsecase  domain.TrafficUsecase

	Publisher 		*kafka.KafkaPublisher
}

func NewDefaultOrderUsecase(
	orderRepo domain.OrderRepository, 
	bankDetailRepo domain.BankDetailRepository, 
	bankingClient *client.BankingClient,
	walletHandler *handlers.HTTPWalletHandler,
	trafficUsecase domain.TrafficUsecase,
	kafkaPublisher *kafka.KafkaPublisher) *DefaultOrderUsecase {

	return &DefaultOrderUsecase{
		OrderRepo: orderRepo,
		BankDetailRepo: bankDetailRepo,
		BankingClient: bankingClient,
		WalletHandler: walletHandler,
		TrafficUsecase: trafficUsecase,
		Publisher: kafkaPublisher,
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
			return nil, err
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
			if order.Status == domain.StatusCreated {
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
		for _, order := range orders {
			if order.Status == domain.StatusSucceed && time.Since(order.UpdatedAt) <= 24*time.Hour {
				ordersSucceedSummary += order.AmountFiat
			}
		}
		fmt.Printf("Max amount a day: %d. Current summary amount: %f\n", bankDetail.MaxAmountDay, ordersSucceedSummary)
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
		for _, order := range orders {
			if order.Status == domain.StatusSucceed && time.Since(order.UpdatedAt) <= 30*24*time.Hour {
				ordersSucceedSummary += order.AmountFiat
			}
		}
		fmt.Printf("Max amount a month: %d. Current summary month: %f\n", bankDetail.MaxAmountMonth, ordersSucceedSummary)
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
			if order.Status != domain.StatusSucceed {
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
			if order.Status == domain.StatusCreated && time.Since(order.CreatedAt) <= 24*time.Hour {
				ordersQuantityDay++
				continue
			}
			if time.Since(order.UpdatedAt) <= 24*time.Hour {
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
		ordersQuantityDay := 0 
		for _, order := range orders {
			if order.Status == domain.StatusCanceled {
				continue
			}
			if order.Status == domain.StatusCreated && time.Since(order.CreatedAt) <= 24*30*time.Hour {
				ordersQuantityDay++
				continue
			}
			if time.Since(order.UpdatedAt) <= 24*30*time.Hour {
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
			if order.Status == domain.StatusCreated && order.AmountFiat == amountFiat {
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

func (uc *DefaultOrderUsecase) FindEligibleBankDetails(order *domain.Order, query *domain.BankDetailQuery) ([]*domain.BankDetail, error) {
	eligibleBankDetailsResponse, err := uc.BankingClient.GetEligibleBankDetails(query)
	if err != nil {
		return nil, err
	}

	if len(eligibleBankDetailsResponse.BankDetails) == 0{
		return nil, domain.ErrNoAvailableBankDetails
	}

	bankDetails := make([]*domain.BankDetail, len(eligibleBankDetailsResponse.BankDetails))
	for i, bankDetail := range eligibleBankDetailsResponse.BankDetails {
		bankDetails[i] = &domain.BankDetail{
			ID: bankDetail.BankDetailId,
			TraderID: bankDetail.TraderId,
			Country: bankDetail.Country,
			Currency: bankDetail.Currency,
			MinAmount: float32(bankDetail.MinAmount),
			MaxAmount: float32(bankDetail.MaxAmount),
			BankName: bankDetail.BankName,
			PaymentSystem: bankDetail.PaymentSystem,
			Delay: bankDetail.Delay.AsDuration(),
			Enabled: bankDetail.Enabled,
			CardNumber: bankDetail.CardNumber,
			Phone: bankDetail.Phone,
			Owner: bankDetail.Owner,
			MaxOrdersSimultaneosly: bankDetail.MaxOrdersSimultaneosly,
			MaxAmountDay: int32(bankDetail.MaxAmountDay),
			MaxAmountMonth: int32(bankDetail.MaxAmountMonth),
			MaxQuantityDay: int32(bankDetail.MaxQuantityDay),
			MaxQuantityMonth: int32(bankDetail.MaxQuantityMonth),
		}
	}

	// 0) Filter by Traffic
	bankDetails, err = uc.FilterByTraffic(bankDetails, order.MerchantID)
	if err != nil {
		return nil, err
	}

	// 1) Filter by Trader Available balances
	bankDetails, err = uc.FilterByTraderBalance(bankDetails, order.AmountCrypto)
	if err != nil {
		return nil, err
	}

	// 2) Filter by MaxOrdersSimultaneosly
	bankDetails, err = uc.FilterByMaxOrdersSimulateosly(bankDetails)
	if err != nil {
		return nil, err
	}
	// 3) Filter by MaxAmountDay
	bankDetails, err = uc.FilterByMaxAmountDay(bankDetails, order.AmountFiat)
	if err != nil {
		return nil, err
	}
	// 4) Filter by MaxAmountMonth
	bankDetails, err = uc.FilterByMaxAmountMonth(bankDetails, order.AmountFiat)
	if err != nil {
		return nil, err
	}
	// 5) Filter by delay
	bankDetails, err = uc.FilterByDelay(bankDetails)
	if err != nil {
		return nil, err
	}
	// 6) Filter by MaxQuantityDay
	bankDetails, err = uc.FilterByMaxQuantityDay(bankDetails)
	if err != nil {
		return nil, err
	}

	// 7) Filter by MaxQuantityMonth
	bankDetails, err = uc.FilterByMaxQuantityMonth(bankDetails)
	if err != nil {
		return nil, err
	}

	// 8) Filter by active order with equal amount fiat
	tempBankDetails, err := uc.FilterByEqualAmountFiat(bankDetails, order.AmountFiat)
	if err != nil {
		return nil, err
	}
	// Если shuffle не задан, то пропускаем сериб проверок с рекалькуляцией
	for addFiat := range order.Shuffle {
		tempBankDetails, err = uc.FilterByEqualAmountFiat(bankDetails, order.AmountFiat + float64(addFiat))
		if err != nil {
			return nil, err
		}
		if len(tempBankDetails) != 0 {
			order.AmountFiat += float64(addFiat)
			if addFiat != 0 {
				order.Recalculated = true
			}else {
				order.Recalculated = false
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

func (uc *DefaultOrderUsecase) CreateOrder(order *domain.Order) (*domain.Order, error) {
	// find eligible bank details
	query := domain.BankDetailQuery{
		Amount: float32(order.AmountFiat),
		Currency: order.Currency,
		PaymentSystem: order.PaymentSystem,
		Country: order.Country,
	}

		// USD RATE
	amountCrypto := float64(order.AmountFiat / usdt.UsdtRubRates)
	order.AmountCrypto = amountCrypto
	order.CryptoRubRate = usdt.UsdtRubRates

	// check idempotency by client_id
	if err := uc.CheckIdempotency(order.ClientID); err != nil {
		return nil, err
	}

	// searching for eligible bank details due to order query parameters
	bankDetails, err := uc.FindEligibleBankDetails(order, &query)
	if err != nil {
		return nil, status.Error(codes.NotFound, "no eligible bank detail")
	}

	// business logic to pick best bank detail
	chosenBankDetail, err := uc.PickBestBankDetail(bankDetails, order.MerchantID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to pick best bank detail gor order: %s", order.ID)
	}

	// relate found bank detail and order
	order.BankDetailsID = chosenBankDetail.ID
	//Save bank detail relevant to order
	if err := uc.BankDetailRepo.SaveBankDetail(chosenBankDetail); err != nil {
		return nil, err
	}

	// Get trader reward percent and save to order
	traffic, err := uc.TrafficUsecase.GetTrafficByTraderMerchant(chosenBankDetail.TraderID, order.MerchantID)
	if err != nil {
		return nil, err
	}
	rewardPercent := traffic.TraderRewardPercent
	platformFee := traffic.PlatformFee
	order.TraderRewardPercent = rewardPercent
	order.PlatformFee = platformFee
	orderID, err := uc.OrderRepo.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	// Freeze crypto
	if err := uc.WalletHandler.Freeze(chosenBankDetail.TraderID, order.ID, amountCrypto); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = uc.Publisher.Publish(kafka.OrderEvent{
		OrderID: order.ID,
		TraderID: chosenBankDetail.TraderID,
		Status: "🔥"+string(order.Status),
		AmountFiat: order.AmountFiat,
		Currency: order.Currency,
	}); err != nil {
		slog.Error("failed to publish event", "error", err.Error())
	}

	return &domain.Order{
		ID: orderID,
		MerchantID: order.MerchantID,
		AmountFiat: order.AmountFiat,
		AmountCrypto: order.AmountCrypto,
		Currency: order.Currency,
		Country: order.Country,
		ClientID: order.ClientID,
		Status: order.Status,
		PaymentSystem: order.PaymentSystem,
		BankDetailsID: order.BankDetailsID,
		ExpiresAt: order.ExpiresAt,
		MerchantOrderID: order.MerchantOrderID,
		Shuffle: order.Shuffle,
		CallbackURL: order.CallbackURL,
		TraderRewardPercent: order.TraderRewardPercent,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
		Recalculated: order.Recalculated,
		CryptoRubRate: order.CryptoRubRate,
		BankDetail: &domain.BankDetail{
			ID: chosenBankDetail.ID,
			TraderID: chosenBankDetail.TraderID,
			Country: chosenBankDetail.Country,
			Currency: chosenBankDetail.Currency,
			MinAmount: chosenBankDetail.MinAmount,
			MaxAmount: chosenBankDetail.MaxAmount,
			BankName: chosenBankDetail.BankName,
			PaymentSystem: chosenBankDetail.PaymentSystem,
			Delay: chosenBankDetail.Delay,
			Enabled: chosenBankDetail.Enabled,
			CardNumber: chosenBankDetail.CardNumber,
			Phone: chosenBankDetail.Phone,
			Owner: chosenBankDetail.Owner,
			MaxOrdersSimultaneosly: chosenBankDetail.MaxOrdersSimultaneosly,
			MaxAmountDay: chosenBankDetail.MaxAmountDay,
			MaxAmountMonth: chosenBankDetail.MaxAmountMonth,
			MaxQuantityDay: chosenBankDetail.MaxQuantityDay,
			MaxQuantityMonth: chosenBankDetail.MaxQuantityMonth,
		},
	}, nil
}

func (uc *DefaultOrderUsecase) GetOrderByID(orderID string) (*domain.Order, error) {
	return uc.OrderRepo.GetOrderByID(orderID)
}

func (uc *DefaultOrderUsecase) GetOrdersByTraderID(
	traderID string, page, 
	limit int64, sortBy, 
	sortOrder string,
	filters domain.OrderFilters,
) ([]*domain.Order, int64, error) {

	validStatuses := map[string]bool{
		"SUCCEED": true,
		"CANCELED": true,
		"CREATED": true,
		"DISPUTE_CREATED": true,
		"DISPUTE_RESOLVED": true,
	}

	for _, status := range filters.Statuses {
		if !validStatuses[status] {
			return nil, 0, fmt.Errorf("invalid status in filters")
		}
	}

	return uc.OrderRepo.GetOrdersByTraderID(
		traderID, 
		page, 
		limit, 
		sortBy, 
		sortOrder,
		filters,
	)
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
		if err := uc.WalletHandler.Release(order.BankDetail.TraderID, order.MerchantID, order.ID, float64(1.), 0); err != nil {
			log.Printf("Unfreeze failed for order %s: %v", order.ID, err)
			return status.Error(codes.Internal, err.Error())
		}
		
		if err := uc.OrderRepo.UpdateOrderStatus(order.ID, domain.StatusCanceled); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		log.Printf("Order %s canceled due to timeout!\n", order.ID)
		// Вызов callback мерчанта
		if(order.CallbackURL != ""){
			notifier.SendCallback(order.CallbackURL, notifier.CallbackPayload{
				OrderID: order.ID,
				MerchantOrderID: order.MerchantOrderID,
				Status: string(domain.StatusCanceled),
				AmountFiat: order.AmountFiat,
				AmountCrypto: order.AmountCrypto,
				Currency: order.Currency,
				ConfirmedAt: order.UpdatedAt,
				ClientID: order.ClientID,
			})
		} 
	}

	return nil
}

func (uc *DefaultOrderUsecase) OpenOrderDispute(orderID string) error {
	// Find exact order
	order, err := uc.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	// Check order status to open dispute (only cancelled can be opened with dispute status)
	if order.Status != domain.StatusCanceled {
		return domain.ErrOpenDisputeFailed
	}

	// Set order status to DISPUTE_CREATED
	if err := uc.OrderRepo.UpdateOrderStatus(orderID, domain.StatusDisputeCreated); err != nil {
		return err
	}

	// Freeze crypto
	fmt.Println("Freezing crypto!")
	fmt.Println(order.AmountCrypto)
	if err := uc.WalletHandler.Freeze(order.BankDetail.TraderID, order.ID, order.AmountCrypto); err != nil {
		return err
	}

	if(order.CallbackURL != ""){
		notifier.SendCallback(order.CallbackURL, notifier.CallbackPayload{
			OrderID: order.ID,
			MerchantOrderID: order.MerchantOrderID,
			Status: string(domain.StatusDisputeCreated),
			AmountFiat: order.AmountFiat,
			AmountCrypto: order.AmountCrypto,
			Currency: order.Currency,
			ConfirmedAt: order.UpdatedAt,
			ClientID: order.ClientID,
		})
	} 

	return nil
}

// Deprecated
func (uc *DefaultOrderUsecase) ResolveOrderDispute(orderID string) error {
	// Find exact order
	order, err := uc.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != domain.StatusDisputeCreated {
		return domain.ErrResolveDisputeFailed
	}


	// Improve
	rewardPercent := order.TraderRewardPercent
	if err := uc.WalletHandler.Release(order.BankDetail.TraderID, order.MerchantID, order.ID, rewardPercent, 0); err != nil {
		return err
	}

	// Set order status to DISPUTE_CREATED
	if err := uc.OrderRepo.UpdateOrderStatus(orderID, domain.StatusDisputeResolved); err != nil {
		return err
	}

	if(order.CallbackURL != ""){
		notifier.SendCallback(order.CallbackURL, notifier.CallbackPayload{
			OrderID: order.ID,
			MerchantOrderID: order.MerchantOrderID,
			Status: string(domain.StatusDisputeResolved),
			AmountFiat: order.AmountFiat,
			AmountCrypto: order.AmountCrypto,
			Currency: order.Currency,
			ConfirmedAt: order.UpdatedAt,
			ClientID: order.ClientID,
		})
	} 

	return nil
}

func (uc *DefaultOrderUsecase) ApproveOrder(orderID string) error {
	// Find exact order
	order, err := uc.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != domain.StatusCreated {
		return domain.ErrResolveDisputeFailed
	}

	rewardPercent := order.TraderRewardPercent
	platformFee := order.PlatformFee
	if err := uc.WalletHandler.Release(order.BankDetail.TraderID, order.MerchantID, order.ID, rewardPercent, platformFee); err != nil {
		return err
	}

	// Set order status to SUCCEED
	if err := uc.OrderRepo.UpdateOrderStatus(orderID, domain.StatusSucceed); err != nil {
		return err
	}

	if err = uc.Publisher.Publish(kafka.OrderEvent{
		OrderID: order.ID,
		TraderID: order.BankDetail.TraderID,
		Status: "✅"+string(domain.StatusSucceed),
		AmountFiat: order.AmountFiat,
		Currency: order.Currency,
	}); err != nil {
		slog.Error("failed to publish event", "error", err.Error())
	}

	// Вызов callback мерчанта
	if(order.CallbackURL != ""){
		notifier.SendCallback(order.CallbackURL, notifier.CallbackPayload{
			OrderID: order.ID,
			MerchantOrderID: order.MerchantOrderID,
			Status: string(domain.StatusSucceed),
			AmountFiat: order.AmountFiat,
			AmountCrypto: order.AmountCrypto,
			Currency: order.Currency,
			ConfirmedAt: order.UpdatedAt,
			ClientID: order.ClientID,
		})
	} 

	return nil
}

func (uc *DefaultOrderUsecase) CancelOrder(orderID string) error {
	// Find exact order
	order, err := uc.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != domain.StatusCreated && order.Status != domain.StatusDisputeCreated{
		return domain.ErrCancelOrder
	}

	// Set order status to CANCELED
	if err := uc.OrderRepo.UpdateOrderStatus(orderID, domain.StatusCanceled); err != nil {
		return err
	}

	if err := uc.WalletHandler.Release(order.BankDetail.TraderID, order.MerchantID, order.ID, 1., 0.); err != nil {
		return err
	}

	if err = uc.Publisher.Publish(kafka.OrderEvent{
		OrderID: order.ID,
		TraderID: order.BankDetail.TraderID,
		Status: "⛔️"+string(domain.StatusCanceled),
		AmountFiat: order.AmountFiat,
		Currency: order.Currency,
	}); err != nil {
		slog.Error("failed to publish event", "error", err.Error())
	}

	// Вызов callback мерчанта
	if(order.CallbackURL != ""){
		notifier.SendCallback(order.CallbackURL, notifier.CallbackPayload{
			OrderID: order.ID,
			MerchantOrderID: order.MerchantOrderID,
			Status: string(domain.StatusCanceled),
			AmountFiat: order.AmountFiat,
			AmountCrypto: order.AmountCrypto,
			Currency: order.Currency,
			ConfirmedAt: order.UpdatedAt,
			ClientID: order.ClientID,
		})
	} 

	return nil
}