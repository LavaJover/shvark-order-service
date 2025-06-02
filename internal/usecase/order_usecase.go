package usecase

import (
	"github.com/LavaJover/shvark-order-service/internal/client"
	"github.com/LavaJover/shvark-order-service/internal/domain"
)

type DefaultOrderUsecase struct {
	OrderRepo 		domain.OrderRepository
	BankDetailRepo 	domain.BankDetailRepository
	BankingClient   *client.BankingClient
}

func NewDefaultOrderUsecase(orderRepo domain.OrderRepository, bankDetailRepo domain.BankDetailRepository, bankingClient *client.BankingClient) *DefaultOrderUsecase {
	return &DefaultOrderUsecase{
		OrderRepo: orderRepo,
		BankDetailRepo: bankDetailRepo,
		BankingClient: bankingClient,
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
		Amount: order.Amount,
		Currency: order.Currency,
		PaymentSystem: order.PaymentSystem,
		Country: order.Country,
	}

	// searching for eligible bank details due to order query parameters
	bankDetails, err := uc.FindEligibleBankDetails(&query)
	if err != nil {
		return nil, err
	}

	// business logic to pick best bank detail
	chosenBankDetail, err := uc.PickBestBankDetail(bankDetails)
	if err != nil {
		return nil, err
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

	return &domain.Order{
		ID: orderID,
		MerchantID: order.MerchantID,
		Amount: order.Amount,
		Currency: order.Currency,
		Country: order.Country,
		ClientEmail: order.ClientEmail,
		MetadataJSON: order.MetadataJSON,
		Status: order.Status,
		PaymentSystem: order.PaymentSystem,
		BankDetailsID: order.BankDetailsID,
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

func (uc *DefaultOrderUsecase) ApproveOrder(orderID string) error {
	return uc.OrderRepo.UpdateOrderStatus(orderID, "COMPLETED")
}

func (uc *DefaultOrderUsecase) CancelOrder(orderID string) error {
	return uc.OrderRepo.UpdateOrderStatus(orderID, "FAILED")
}

func (uc *DefaultOrderUsecase) GetOrderByID(orderID string) (*domain.Order, error) {
	return uc.OrderRepo.GetOrderByID(orderID)
}

func (uc *DefaultOrderUsecase) GetOrdersByTraderID(traderID string) ([]*domain.Order, error) {
	return uc.OrderRepo.GetOrdersByTraderID(traderID)
}