package grpcapi

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/client"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/protobuf/types/known/durationpb"
)

type OrderHandler struct {
	uc domain.OrderUsecase
	client *client.BankingClient
	orderpb.UnimplementedOrderServiceServer
}

func NewOrderHandler(uc domain.OrderUsecase, client *client.BankingClient) *OrderHandler {
	return &OrderHandler{
		uc: uc,
		client: client,
	}
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
		BankDetailsID: chosenBankDetail.ID,
	}
	
	orderID, err := h.uc.CreateOrder(&order)
	if err != nil {
		return nil, err
	}

	return &orderpb.CreateOrderResponse{
		Order: &orderpb.Order{
			OrderId: orderID,
			Status: orderpb.OrderStatus_DETAILS_PROVIDED,
			BankDetail: &orderpb.BankDetail{
				BankDetailId: chosenBankDetail.ID,
				TraderId: chosenBankDetail.TraderID,
				Currency: chosenBankDetail.Currency,
				Country: chosenBankDetail.Country,
				MinAmount: float64(chosenBankDetail.MinAmount),
				MaxAmount: float64(chosenBankDetail.MaxAmount),
				BankName: chosenBankDetail.BankName,
				PaymentSystem: chosenBankDetail.PaymentSystem,
				Enabled: chosenBankDetail.Enabled,
				Delay: durationpb.New(chosenBankDetail.Delay),
			},
			Amount: float64(order.Amount),
		},
	}, nil
}

func (h *OrderHandler) ApproveOrder(ctx context.Context, r *orderpb.ApproveOrderRequest) (*orderpb.ApproveOrderResponse, error) {
	orderID := r.OrderId
	if err := h.uc.ApproveOrder(orderID); err != nil {
		return nil, err
	}

	return &orderpb.ApproveOrderResponse{
		Message: "Order was successfully approved",
	}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, r *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	orderID := r.OrderId
	if err := h.uc.CancelOrder(orderID); err != nil {
		return nil, err
	}

	return &orderpb.CancelOrderResponse{
		Message: "Order was successfully cancelled",
	}, nil
}

func (h *OrderHandler) GetOrderByID(ctx context.Context, r *orderpb.GetOrderByIDRequest) (*orderpb.GetOrderByIDResponse, error) {
	orderID := r.OrderId
	responseOrder, err := h.uc.GetOrderByID(orderID)
	if err != nil {
		return nil, err
	}

	// get corresponding bank detail for order
	bankDetailResponse, err := h.client.GetBankDetailByID(responseOrder.BankDetailsID)
	if err != nil {
		return nil, err
	}

	return &orderpb.GetOrderByIDResponse{
		Order: &orderpb.Order{
			OrderId: responseOrder.ID,
			Status: orderpb.OrderStatus(orderpb.OrderStatus_value[responseOrder.Status]),
			BankDetail: &orderpb.BankDetail{
				BankDetailId: bankDetailResponse.BankDetail.BankDetailId,
				TraderId: bankDetailResponse.BankDetail.TraderId,
				Currency: bankDetailResponse.BankDetail.Currency,
				Country: bankDetailResponse.BankDetail.Country,
				MinAmount: bankDetailResponse.BankDetail.MinAmount,
				MaxAmount: bankDetailResponse.BankDetail.MaxAmount,
				BankName: bankDetailResponse.BankDetail.BankName,
				PaymentSystem: bankDetailResponse.BankDetail.PaymentSystem,
				Enabled: bankDetailResponse.BankDetail.Enabled,
				Delay: bankDetailResponse.BankDetail.Delay,
			},
			Amount: float64(responseOrder.Amount),
		},
	}, nil
}

func (h *OrderHandler) GetOrdersByTraderID(ctx context.Context, r *orderpb.GetOrdersByTraderIDRequest) (*orderpb.GetOrdersByTraderIDResponse, error) {
	traderID := r.TraderId
	responseOrders, err := h.uc.GetOrdersByTraderID(traderID)
	if err != nil {
		return nil, err
	}

	orders := make([]*orderpb.Order, len(responseOrders))
	for i, responseOrder := range responseOrders {
		orders[i] = &orderpb.Order{
			OrderId: responseOrder.ID,
			Status: orderpb.OrderStatus(orderpb.OrderStatus_value[responseOrder.Status]),
			BankDetail: &orderpb.BankDetail{},
			Amount: float64(responseOrder.Amount),
		}
	}

	return &orderpb.GetOrdersByTraderIDResponse{
		Orders: orders,
	}, nil
}