package grpcapi

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/client"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
)

type OrderHandler struct {
	uc domain.OrderUsecase
	client client.BankingClient
}

func NewOrderHandler(uc domain.OrderUsecase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, r *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	// create query
	query := domain.BankDetailQuery{
		Amount: float32(r.Amount),
		Currency: r.Currency,
		PaymentSystem: r.PaymentSystem,
		Country: r.Country,
	}
	// find bank details
	response, err := h.client.GetEligibleBankDetails(&query)
	if err != nil {
		return nil, err
	}

	bankDetailResponse := response.BankDetails

	if len(bankDetailResponse) == 0{
		return nil, domain.ErrNoAvailableBankDetails
	}

	bankDetails := make([]*domain.BankDetail, len(bankDetailResponse))
	for i, bankDetail := range bankDetailResponse {
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

	// logic for choosing bankDetail, load-balancing
	chosenBankDetail := bankDetails[0]
	// 
	order := domain.Order{
		MerchantID: r.MerchantId,
		Amount: float32(r.Amount),
		Currency: r.Currency,
		Country: r.Country,
		ClientEmail: r.ClientEmail,
		MetadataJSON: r.MetadataJson,
		Status: orderpb.OrderStatus_DETAILS_PROVIDED.String(),
		PaymentSystem: r.PaymentSystem,
	}
	
	orderID, err := h.uc.CreateOrder(&order)
	if err != nil {
		return nil, err
	}

	return &orderpb.CreateOrderResponse{
		OrderId: orderID,
		Status: orderpb.OrderStatus_DETAILS_PROVIDED,
		BankDetailId: chosenBankDetail.ID,
	}, nil
}