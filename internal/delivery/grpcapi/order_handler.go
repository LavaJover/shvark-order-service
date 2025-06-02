package grpcapi

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/protobuf/types/known/durationpb"
)

type OrderHandler struct {
	uc domain.OrderUsecase
	orderpb.UnimplementedOrderServiceServer
}

func NewOrderHandler(uc domain.OrderUsecase) *OrderHandler {
	return &OrderHandler{
		uc: uc,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, r *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {

	orderRequest := domain.Order{
		MerchantID: r.MerchantId,
		Amount: float32(r.Amount),
		Currency: r.Currency,
		Country: r.Country,
		ClientEmail: r.ClientEmail,
		MetadataJSON: r.MetadataJson,
		Status: orderpb.OrderStatus_DETAILS_PROVIDED.String(),
		PaymentSystem: r.PaymentSystem,
	}
	
	savedOrder, err := h.uc.CreateOrder(&orderRequest)
	if err != nil {
		return nil, err
	}

	return &orderpb.CreateOrderResponse{
		Order: &orderpb.Order{
			OrderId: savedOrder.ID,
			Status: orderpb.OrderStatus_DETAILS_PROVIDED,
			BankDetail: &orderpb.BankDetail{
				BankDetailId: savedOrder.BankDetail.ID,
				TraderId: savedOrder.BankDetail.TraderID,
				Currency: savedOrder.BankDetail.Currency,
				Country: savedOrder.BankDetail.Country, 
				MinAmount: float64(savedOrder.BankDetail.MinAmount),
				MaxAmount: float64(savedOrder.BankDetail.MaxAmount),
				BankName: savedOrder.BankDetail.BankName,
				PaymentSystem: savedOrder.BankDetail.PaymentSystem,
				Enabled: savedOrder.BankDetail.Enabled,
				Delay: durationpb.New(savedOrder.BankDetail.Delay),
			},
			Amount: float64(savedOrder.Amount),
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
	orderResponse, err := h.uc.GetOrderByID(orderID)
	if err != nil {
		return nil, err
	}

	return &orderpb.GetOrderByIDResponse{
		Order: &orderpb.Order{
			OrderId: orderID,
			Status: orderpb.OrderStatus(orderpb.OrderStatus_value[orderResponse.Status]),
			BankDetail: &orderpb.BankDetail{
				BankDetailId: orderResponse.BankDetail.ID,
				TraderId: orderResponse.BankDetail.TraderID,
				Currency: orderResponse.BankDetail.Currency,
				Country: orderResponse.BankDetail.Country,
				MinAmount: float64(orderResponse.BankDetail.MinAmount),
				MaxAmount: float64(orderResponse.BankDetail.MaxAmount),
				BankName: orderResponse.BankDetail.BankName,
				PaymentSystem: orderResponse.BankDetail.PaymentSystem,
				Enabled: orderResponse.BankDetail.Enabled,
				Delay: durationpb.New(orderResponse.BankDetail.Delay),
			},
			Amount: float64(orderResponse.Amount),
		},
	}, nil
}

func (h *OrderHandler) GetOrdersByTraderID(ctx context.Context, r *orderpb.GetOrdersByTraderIDRequest) (*orderpb.GetOrdersByTraderIDResponse, error) {
	return nil, nil
}