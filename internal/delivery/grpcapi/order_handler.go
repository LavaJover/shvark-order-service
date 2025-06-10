package grpcapi

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		AmountFiat: float32(r.AmountFiat),
		Currency: r.Currency,
		Country: r.Country,
		ClientEmail: r.ClientEmail,
		MetadataJSON: r.MetadataJson,
		Status: domain.StatusCreated,
		PaymentSystem: r.PaymentSystem,
		ExpiresAt: r.ExpiresAt.AsTime(),
	}
	
	savedOrder, err := h.uc.CreateOrder(&orderRequest)
	if err != nil {
		return nil, err
	}

	return &orderpb.CreateOrderResponse{
		Order: &orderpb.Order{
			OrderId: savedOrder.ID,
			Status: string(savedOrder.Status),
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
			AmountFiat: float64(savedOrder.AmountFiat),
			AmountCrypto: float64(savedOrder.AmountCrypto),
			ExpiresAt: timestamppb.New(savedOrder.ExpiresAt),
		},
	}, nil
}

func (h *OrderHandler) ApproveOrder(ctx context.Context, r *orderpb.ApproveOrderRequest) (*orderpb.ApproveOrderResponse, error) {
	orderID := r.OrderId
	if err := h.uc.ApproveOrder(orderID); err != nil {
		return &orderpb.ApproveOrderResponse{
			Message: "failed to approve order",
		}, err
	}else {
		return &orderpb.ApproveOrderResponse{
			Message: "Order approved successfully",
		}, nil
	}
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
			Status: string(orderResponse.Status),
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
			AmountFiat: float64(orderResponse.AmountFiat),
			AmountCrypto: float64(orderResponse.AmountCrypto),
			ExpiresAt: timestamppb.New(orderResponse.ExpiresAt),
		},
	}, nil
}

func (h *OrderHandler) GetOrdersByTraderID(ctx context.Context, r *orderpb.GetOrdersByTraderIDRequest) (*orderpb.GetOrdersByTraderIDResponse, error) {
	traderID := r.TraderId
	ordersResponse, err := h.uc.GetOrdersByTraderID(traderID)
	if err != nil {
		return nil, err
	}

	orders := make([]*orderpb.Order, len(ordersResponse))
	for i, order := range ordersResponse {
		orders[i] = &orderpb.Order{
			OrderId: order.ID,
			Status: string(order.Status),
			BankDetail: &orderpb.BankDetail{
				BankDetailId: order.BankDetail.ID,
				TraderId: order.BankDetail.TraderID,
				Currency: order.BankDetail.Currency,
				Country: order.BankDetail.Country,
				MinAmount: float64(order.BankDetail.MinAmount),
				MaxAmount: float64(order.BankDetail.MaxAmount),
				BankName: order.BankDetail.BankName,
				PaymentSystem: order.BankDetail.PaymentSystem,
				Enabled: order.BankDetail.Enabled,
				Delay: durationpb.New(order.BankDetail.Delay),
			},
			AmountFiat: float64(order.AmountFiat),
			AmountCrypto: float64(order.AmountCrypto),
			ExpiresAt: timestamppb.New(order.ExpiresAt),
		}
	}

	return &orderpb.GetOrdersByTraderIDResponse{
		Orders: orders,
	}, nil

}

func (h *OrderHandler) OpenOrderDispute(ctx context.Context, r *orderpb.OpenOrderDisputeRequest) (*orderpb.OpenOrderDisputeResponse, error) {
	orderID := r.OrderId
	if err := h.uc.OpenOrderDispute(orderID); err != nil {
		return &orderpb.OpenOrderDisputeResponse{
			Message: "Failed to open dispute: " + err.Error(),
		}, err
	}else {
		return &orderpb.OpenOrderDisputeResponse{
			Message: "Dispute opened successfully!",
		}, nil
	}
}

func (h *OrderHandler) ResolveOrderDispute(ctx context.Context, r *orderpb.ResolveOrderDisputeRequest) (*orderpb.ResolveOrderDisputeResponse, error) {
	orderID := r.OrderId
	if err := h.uc.ResolveOrderDispute(orderID); err != nil {
		return &orderpb.ResolveOrderDisputeResponse{
			Message: "Failed to resolve dispute: " + err.Error(),
		}, err
	}else {
		return &orderpb.ResolveOrderDisputeResponse{
			Message: "Dispute resolved successfully!",
		}, nil
	}
}