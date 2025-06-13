package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/client"
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DefaultOrderUsecase struct {
	OrderRepo 		domain.OrderRepository
	BankDetailRepo 	domain.BankDetailRepository
	BankingClient   *client.BankingClient
	WalletHandler   *handlers.HTTPWalletHandler
}

func NewDefaultOrderUsecase(
	orderRepo domain.OrderRepository, 
	bankDetailRepo domain.BankDetailRepository, 
	bankingClient *client.BankingClient,
	walletHandler *handlers.HTTPWalletHandler) *DefaultOrderUsecase {

	return &DefaultOrderUsecase{
		OrderRepo: orderRepo,
		BankDetailRepo: bankDetailRepo,
		BankingClient: bankingClient,
		WalletHandler: walletHandler,
	}
}

func (uc *DefaultOrderUsecase) PickBestBankDetail(bankDetails []*domain.BankDetail) (*domain.BankDetail, error) {
	return bankDetails[0], nil
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
		fmt.Printf("Max amount a month: %d. Current summary month: %f\n", bankDetail.MaxAmountDay, ordersSucceedSummary)
		if ordersSucceedSummary + amountFiat <= float64(bankDetail.MaxAmountDay) {
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

	// 1) Filter by MaxOrdersSimultaneosly
	bankDetails, err = uc.FilterByMaxOrdersSimulateosly(bankDetails)
	if err != nil {
		return nil, err
	}
	// 2) Filter by MaxAmountDay
	bankDetails, err = uc.FilterByMaxAmountDay(bankDetails, order.AmountFiat)
	if err != nil {
		return nil, err
	}
	// 3) Filter by MaxAmountMonth
	bankDetails, err = uc.FilterByMaxAmountMonth(bankDetails, order.AmountFiat)
	if err != nil {
		return nil, err
	}
	// 4) Filter by delay
	bankDetails, err = uc.FilterByDelay(bankDetails)
	if err != nil {
		return nil, err
	}
	// 5) Filter by MaxQuantityDay
	bankDetails, err = uc.FilterByMaxQuantityDay(bankDetails)
	if err != nil {
		return nil, err
	}

	if len(bankDetails) == 0 {
		return nil, domain.ErrNoAvailableBankDetails
	}

	return bankDetails, nil
}

func (uc *DefaultOrderUsecase) CreateOrder(order *domain.Order) (*domain.Order, error) {
	// find eligible bank details
	query := domain.BankDetailQuery{
		Amount: float32(order.AmountFiat),
		Currency: order.Currency,
		PaymentSystem: order.PaymentSystem,
		Country: order.Country,
	}

	// searching for eligible bank details due to order query parameters
	bankDetails, err := uc.FindEligibleBankDetails(order, &query)
	if err != nil {
		return nil, status.Error(codes.NotFound, "no eligible bank detail")
	}

	// business logic to pick best bank detail
	chosenBankDetail, err := uc.PickBestBankDetail(bankDetails)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to pick best bank detail")
	}

	fmt.Println(chosenBankDetail)

	// relate found bank detail and order
	order.BankDetailsID = chosenBankDetail.ID

	//Save bank detail relevant to order
	if err := uc.BankDetailRepo.SaveBankDetail(chosenBankDetail); err != nil {
		return nil, err
	}

		// BTC RATE
	// IMPROVE THIS !!!
	amountCrypto := float64(order.AmountFiat / 8599022)
	order.AmountCrypto = amountCrypto
	////////////////////////////////////////////////

	orderID, err := uc.OrderRepo.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	fmt.Println(chosenBankDetail.TraderID, order.ID, amountCrypto)

	// Freeze crypto
	if err := uc.WalletHandler.Freeze(chosenBankDetail.TraderID, order.ID, amountCrypto); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &domain.Order{
		ID: orderID,
		MerchantID: order.MerchantID,
		AmountFiat: order.AmountFiat,
		AmountCrypto: order.AmountCrypto,
		Currency: order.Currency,
		Country: order.Country,
		ClientEmail: order.ClientEmail,
		MetadataJSON: order.MetadataJSON,
		Status: order.Status,
		PaymentSystem: order.PaymentSystem,
		BankDetailsID: order.BankDetailsID,
		ExpiresAt: order.ExpiresAt,
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

func (uc *DefaultOrderUsecase) GetOrdersByTraderID(traderID string) ([]*domain.Order, error) {
	return uc.OrderRepo.GetOrdersByTraderID(traderID)
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
		if err := uc.WalletHandler.Release(order.BankDetail.TraderID, order.ID, float64(0.)); err != nil {
			log.Printf("Unfreeze failed for order %s: %v", order.ID, err)
			return status.Error(codes.Internal, err.Error())
		}
		
		if err := uc.OrderRepo.UpdateOrderStatus(order.ID, domain.StatusCanceled); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		log.Printf("Order %s canceled due to timeout!\n", order.ID)
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

	return nil
}

func (uc *DefaultOrderUsecase) ResolveOrderDispute(orderID string) error {
	// Find exact order
	order, err := uc.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != domain.StatusDisputeCreated {
		return domain.ErrResolveDisputeFailed
	}

	// Set order status to DISPUTE_CREATED
	if err := uc.OrderRepo.UpdateOrderStatus(orderID, domain.StatusDisputeResolved); err != nil {
		return err
	}

	// Improve
	rewardPercent := float64(0.09)
	if err := uc.WalletHandler.Release(order.BankDetail.TraderID, order.ID, rewardPercent); err != nil {
		return err
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

	// Set order status to DISPUTE_CREATED
	if err := uc.OrderRepo.UpdateOrderStatus(orderID, domain.StatusSucceed); err != nil {
		return err
	}

	// Improve
	rewardPercent := float64(0.09)
	if err := uc.WalletHandler.Release(order.BankDetail.TraderID, order.ID, rewardPercent); err != nil {
		return err
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

	// Set order status to DISPUTE_CREATED
	if err := uc.OrderRepo.UpdateOrderStatus(orderID, domain.StatusCanceled); err != nil {
		return err
	}

	if err := uc.WalletHandler.Release(order.BankDetail.TraderID, order.ID, 0); err != nil {
		return err
	}

	return nil
}