package usecase

import (
	"fmt"
	"log"

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

func (uc *DefaultOrderUsecase) FindEligibleBankDetails(query *domain.BankDetailQuery) ([]*domain.BankDetail, error) {
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
		}
	}

	return bankDetails, nil
	
}

func (uc *DefaultOrderUsecase) CreateOrder(order *domain.Order) (*domain.Order, error) {
	// find eligible bank details
	query := domain.BankDetailQuery{
		Amount: order.AmountFiat,
		Currency: order.Currency,
		PaymentSystem: order.PaymentSystem,
		Country: order.Country,
	}

	// searching for eligible bank details due to order query parameters
	bankDetails, err := uc.FindEligibleBankDetails(&query)
	if err != nil {
		return nil, status.Error(codes.NotFound, "no eligible bank detail")
	}

	// business logic to pick best bank detail
	chosenBankDetail, err := uc.PickBestBankDetail(bankDetails)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to pick best bank detail")
	}

	// relate found bank detail and order
	order.BankDetailsID = chosenBankDetail.ID

	//Save bank detail relevant to order
	if err := uc.BankDetailRepo.SaveBankDetail(chosenBankDetail); err != nil {
		return nil, err
	}

	orderID, err := uc.OrderRepo.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	// BTC RATE
	// IMPROVE THIS !!!
	amountCrypto := float64(order.AmountFiat / 8599022)

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

func (uc *DefaultOrderUsecase) CancelExpiredOrders() error {
	orders, err := uc.FindExpiredOrders()
	if err != nil {
		return nil
	}

	for _, order := range orders {
		if err := uc.WalletHandler.Release(order.BankDetail.TraderID, order.ID, 1); err != nil {
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

func (uc *DefaultOrderUsecase) UpdateOrderStatus(orderID string, newStatus domain.OrderStatus) error {
	return uc.OrderRepo.UpdateOrderStatus(orderID, newStatus)
}