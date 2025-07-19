package grpcapi

import (
	"context"
	"math"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderHandler struct {
	uc domain.OrderUsecase
	disputeUc domain.DisputeUsecase
	orderpb.UnimplementedOrderServiceServer
}

func NewOrderHandler(
	uc domain.OrderUsecase,
	disputeUc domain.DisputeUsecase,
	) *OrderHandler {
	return &OrderHandler{
		uc: uc,
		disputeUc: disputeUc,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, r *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {

	orderRequest := domain.Order{
		MerchantID: r.MerchantId,
		AmountFiat: r.AmountFiat,
		Currency: r.Currency,
		Country: r.Country,
		ClientID: r.ClientId,
		Status: domain.StatusCreated,
		PaymentSystem: r.PaymentSystem,
		MerchantOrderID: r.MerchantOrderId,
		Shuffle: r.Shuffle,
		ExpiresAt: r.ExpiresAt.AsTime(),
		CallbackURL: r.CallbackUrl,
		BankCode: r.BankCode,
		NspkCode: r.NspkCode,
	}
	
	savedOrder, err := h.uc.CreateOrder(&orderRequest)
	if err != nil {
		if orderRequest.CallbackURL != ""{
			notifier.SendCallback(
				orderRequest.CallbackURL,
				orderRequest.MerchantOrderID,
				string(domain.StatusFailed),
				0, 0, 0,
			)
		}
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
				Owner: savedOrder.BankDetail.Owner,
				CardNumber: savedOrder.BankDetail.CardNumber,
				Phone: savedOrder.BankDetail.Phone,
				MaxOrdersSimultaneosly: savedOrder.BankDetail.MaxOrdersSimultaneosly,
				MaxAmountDay: float64(savedOrder.BankDetail.MaxAmountDay),
				MaxAmountMonth: float64(savedOrder.BankDetail.MaxAmountMonth),
				MaxQuantityDay: float64(savedOrder.BankDetail.MaxQuantityDay),
				MaxQuantityMonth: float64(savedOrder.BankDetail.MaxQuantityMonth),
				DeviceId: savedOrder.BankDetail.DeviceID,
				InflowCurrency: savedOrder.BankDetail.InflowCurrency,
				BankCode: savedOrder.BankDetail.BankCode,
				NspkCode: savedOrder.BankDetail.NspkCode,
			},
			AmountFiat: float64(savedOrder.AmountFiat),
			AmountCrypto: float64(savedOrder.AmountCrypto),
			ExpiresAt: timestamppb.New(savedOrder.ExpiresAt),
			Shuffle: savedOrder.Shuffle,
			MerchantOrderId: savedOrder.MerchantOrderID,
			ClientId: savedOrder.ClientID,
			CallbackUrl: savedOrder.CallbackURL,
			TraderRewardPercent: savedOrder.TraderRewardPercent,
			CreatedAt: timestamppb.New(savedOrder.CreatedAt),
			UpdatedAt: timestamppb.New(savedOrder.UpdatedAt),
			Recalculated: savedOrder.Recalculated,
			CryptoRubRate: savedOrder.CryptoRubRate,
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
				Owner: orderResponse.BankDetail.Owner,
				CardNumber: orderResponse.BankDetail.CardNumber,
				Phone: orderResponse.BankDetail.Phone,
				BankCode: orderResponse.BankDetail.BankCode,
				NspkCode: orderResponse.BankDetail.NspkCode,
				InflowCurrency: orderResponse.BankDetail.InflowCurrency,
			},
			AmountFiat: float64(orderResponse.AmountFiat),
			AmountCrypto: float64(orderResponse.AmountCrypto),
			ExpiresAt: timestamppb.New(orderResponse.ExpiresAt),
			MerchantOrderId: orderResponse.MerchantOrderID,
			Shuffle: orderResponse.Shuffle,
			ClientId: orderResponse.ClientID,
			CallbackUrl: orderResponse.CallbackURL,
			TraderRewardPercent: orderResponse.TraderRewardPercent,
			CreatedAt: timestamppb.New(orderResponse.CreatedAt),
			UpdatedAt: timestamppb.New(orderResponse.UpdatedAt),
			Recalculated: orderResponse.Recalculated,
			CryptoRubRate: orderResponse.CryptoRubRate,
		},
	}, nil
}

func (h *OrderHandler) GetOrderByMerchantOrderID(ctx context.Context, r *orderpb.GetOrderByMerchantOrderIDRequest) (*orderpb.GetOrderByMerchantOrderIDResponse, error) {
	merchantOrderID := r.MerchantOrderId
	orderResponse, err := h.uc.GetOrderByMerchantOrderID(merchantOrderID)
	if err != nil {
		return nil, err
	}

	return &orderpb.GetOrderByMerchantOrderIDResponse{
		Order: &orderpb.Order{
			OrderId: orderResponse.ID,
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
				Owner: orderResponse.BankDetail.Owner,
				CardNumber: orderResponse.BankDetail.CardNumber,
				Phone: orderResponse.BankDetail.Phone,
				BankCode: orderResponse.BankDetail.BankCode,
				NspkCode: orderResponse.BankDetail.NspkCode,
				InflowCurrency: orderResponse.BankDetail.InflowCurrency,
			},
			AmountFiat: float64(orderResponse.AmountFiat),
			AmountCrypto: float64(orderResponse.AmountCrypto),
			ExpiresAt: timestamppb.New(orderResponse.ExpiresAt),
			MerchantOrderId: orderResponse.MerchantOrderID,
			Shuffle: orderResponse.Shuffle,
			ClientId: orderResponse.ClientID,
			CallbackUrl: orderResponse.CallbackURL,
			TraderRewardPercent: orderResponse.TraderRewardPercent,
			CreatedAt: timestamppb.New(orderResponse.CreatedAt),
			UpdatedAt: timestamppb.New(orderResponse.UpdatedAt),
			Recalculated: orderResponse.Recalculated,
			CryptoRubRate: orderResponse.CryptoRubRate,
		},
	}, nil
}

func (h *OrderHandler) GetOrdersByTraderID(ctx context.Context, r *orderpb.GetOrdersByTraderIDRequest) (*orderpb.GetOrdersByTraderIDResponse, error) {
	// sort_by validation
	validSortFields := map[string]bool{
		"amount_fiat": true,
		"expires_at": true,
		"created_at": true,
	}
	if !validSortFields[r.GetSortBy()] {
		r.SortBy = "created_at"
	}
	// sort_order validation
	if r.GetSortOrder() != "asc" && r.GetSortOrder() != "desc" {
		r.SortOrder = "desc"
	}

	var dateFrom, dateTo time.Time
	if r.Filters.DateFrom != nil {
		dateFrom = r.Filters.DateFrom.AsTime()
	}
	if r.Filters.DateFrom != nil {
		dateTo = r.Filters.DateTo.AsTime()
	}
	filters := domain.OrderFilters{
		Statuses: r.Filters.Statuses,
		MinAmountFiat: r.Filters.MinAmountFiat,
		MaxAmountFiat: r.Filters.MaxAmountFiat,
		DateFrom: dateFrom,
		DateTo: dateTo,
		Currency: r.Filters.Currency,
	}

	traderID := r.TraderId
	ordersResponse, total, err := h.uc.GetOrdersByTraderID(
		traderID, 
		r.GetPage(), 
		r.GetLimit(), 
		r.GetSortBy(), 
		r.GetSortOrder(),
		filters,
	)
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
				Owner: order.BankDetail.Owner,
				CardNumber: order.BankDetail.CardNumber,
				Phone: order.BankDetail.Phone,
				BankCode: order.BankDetail.BankCode,
				NspkCode: order.BankDetail.NspkCode,
			},
			AmountFiat: float64(order.AmountFiat),
			AmountCrypto: float64(order.AmountCrypto),
			ExpiresAt: timestamppb.New(order.ExpiresAt),
			MerchantOrderId: order.MerchantOrderID,
			Shuffle: order.Shuffle,
			ClientId: order.ClientID,
			CallbackUrl: order.CallbackURL,
			TraderRewardPercent: order.TraderRewardPercent,
			CreatedAt: timestamppb.New(order.CreatedAt),
			UpdatedAt: timestamppb.New(order.UpdatedAt),
			Recalculated: order.Recalculated,
			CryptoRubRate: order.CryptoRubRate,
		}
	}

	return &orderpb.GetOrdersByTraderIDResponse{
		Orders: orders,
		Pagination: &orderpb.Pagination{
			CurrentPage: r.Page,
			TotalPages:  int64(math.Ceil(float64(total) / float64(r.Limit))),
			TotalItems: total,
			ItemsPerPage: r.Limit,
		},
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

func (h *OrderHandler) CancelOrder(ctx context.Context, r *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	orderID := r.OrderId
	if err := h.uc.CancelOrder(orderID); err != nil {
		return &orderpb.CancelOrderResponse{
			Message: "Failed to cancel order",
		}, err
	}

	return &orderpb.CancelOrderResponse{
		Message: "Order successfully canceled",
	}, nil
}

func (h *OrderHandler) CreateOrderDispute(ctx context.Context, r *orderpb.CreateOrderDisputeRequest) (*orderpb.CreateOrderDisputeResponse, error) {
	dispute := &domain.Dispute{
		OrderID: r.OrderId,
		ProofUrl: r.ProofUrl,
		Reason: r.DisputeReason,
		Ttl: r.Ttl.AsDuration(),
		DisputeAmountFiat: float64(r.DisputeAmountFiat),
	}
	if err := h.disputeUc.CreateDispute(dispute); err != nil {
		return nil, err
	}
	return &orderpb.CreateOrderDisputeResponse{
		DisputeId: dispute.ID,
	}, nil
}

func (h *OrderHandler) AcceptOrderDispute(ctx context.Context, r *orderpb.AcceptOrderDisputeRequest) (*orderpb.AcceptOrderDisputeResponse, error) {
	disputeID := r.DisputeId
	if err := h.disputeUc.AcceptDispute(disputeID); err != nil {
		return nil, err
	}

	return &orderpb.AcceptOrderDisputeResponse{
		Message: "dispute accepted",
	}, nil
}

func (h *OrderHandler) RejectOrderDispute(ctx context.Context, r *orderpb.RejectOrderDisputeRequest) (*orderpb.RejectOrderDisputeResponse, error) {
	disputeID := r.DisputeId
	if err := h.disputeUc.RejectDispute(disputeID); err != nil {
		return nil, err
	}

	return &orderpb.RejectOrderDisputeResponse{
		Message: "dispute rejected",
	}, nil
}

func (h *OrderHandler) GetOrderDisputeInfo(ctx context.Context, r *orderpb.GetOrderDisputeInfoRequest) (*orderpb.GetOrderDisputeInfoResponse, error) {
	dispueID := r.DisputeId
	dispute, err := h.disputeUc.GetDisputeByID(dispueID)
	if err != nil {
		return nil, err
	}

	return &orderpb.GetOrderDisputeInfoResponse{
		Dispute: &orderpb.OrderDispute{
			DisputeId: dispute.ID,
			OrderId: dispute.OrderID,
			ProofUrl: dispute.ProofUrl,
			DisputeReason: dispute.Reason,
			DisputeStatus: string(dispute.Status),
			DisputeAmountFiat: dispute.DisputeAmountFiat,
			DisputeAmountCrypto: dispute.DisputeAmountCrypto,
			DisputeCryptoRate: dispute.DisputeCryptoRate,
			AcceptAt: timestamppb.New(dispute.AutoAcceptAt),
		},
	}, nil
}

func (h *OrderHandler) FreezeOrderDispute(ctx context.Context, r *orderpb.FreezeOrderDisputeRequest) (*orderpb.FreezeOrderDisputeResponse, error) {
	disputeID := r.DisputeId
	err := h.disputeUc.FreezeDispute(disputeID)
	if err != nil {
		return nil, err
	}
	return &orderpb.FreezeOrderDisputeResponse{}, nil
}

func (h *OrderHandler) GetOrderDisputes(ctx context.Context, r *orderpb.GetOrderDisputesRequest) (*orderpb.GetOrderDisputesResponse, error) {
	page, limit, status := r.Page, r.Limit, r.Status
	disputes, total, err := h.disputeUc.GetOrderDisputes(page, limit, status)
	if err != nil {
		return nil, err
	}

	disputesResp := make([]*orderpb.OrderDispute, len(disputes))
	for i, dispute := range disputes {
		order, err := h.uc.GetOrderByID(dispute.OrderID)
		if err != nil {
			return nil, err
		}
		disputesResp[i] = &orderpb.OrderDispute{
			DisputeId: dispute.ID,
			OrderId: dispute.OrderID,
			ProofUrl: dispute.ProofUrl,
			DisputeReason: dispute.Reason,
			DisputeStatus: string(dispute.Status),
			DisputeAmountFiat: dispute.DisputeAmountFiat,
			DisputeAmountCrypto: dispute.DisputeAmountCrypto,
			DisputeCryptoRate: dispute.DisputeCryptoRate,
			AcceptAt: timestamppb.New(dispute.AutoAcceptAt),
			Order: &orderpb.Order{
				OrderId: order.ID,
				Status: string(order.Status),
				AmountFiat: order.AmountFiat,
				AmountCrypto: order.AmountCrypto,
				ExpiresAt: timestamppb.New(order.ExpiresAt),
				MerchantOrderId: order.MerchantID,
				TraderRewardPercent: order.TraderRewardPercent,
				CreatedAt: timestamppb.New(order.CreatedAt),
				UpdatedAt: timestamppb.New(order.UpdatedAt),
				CryptoRubRate: order.CryptoRubRate,
				BankDetail: &orderpb.BankDetail{
					BankDetailId: order.BankDetail.ID,
					TraderId: order.BankDetail.TraderID,
					Currency: order.BankDetail.Currency,
					BankName: order.BankDetail.BankName,
					PaymentSystem: order.BankDetail.PaymentSystem,
					CardNumber: order.BankDetail.CardNumber,
					Phone: order.BankDetail.Phone,
					Owner: order.BankDetail.Owner,
					DeviceId: order.BankDetail.DeviceID,
					BankCode: order.BankDetail.BankCode,
					Country: order.BankDetail.Country,
					MinAmount: float64(order.BankDetail.MinAmount),
					MaxAmount: float64(order.BankDetail.MaxAmount),
					Enabled: order.BankDetail.Enabled,
					Delay: durationpb.New(order.BankDetail.Delay),
					MaxOrdersSimultaneosly: order.BankDetail.MaxOrdersSimultaneosly,
					MaxAmountDay: float64(order.BankDetail.MaxAmountDay),
					MaxAmountMonth: float64(order.BankDetail.MaxAmountMonth),
					MaxQuantityDay: float64(order.BankDetail.MaxQuantityDay),
					MaxQuantityMonth: float64(order.BankDetail.MaxQuantityMonth),
					InflowCurrency: order.BankDetail.InflowCurrency,
					NspkCode: order.BankDetail.NspkCode,
				},
			},
		}
	}

	return &orderpb.GetOrderDisputesResponse{
		Disputes: disputesResp,
		Pagination: &orderpb.Pagination{
			CurrentPage: r.Page,
			TotalPages:  int64(math.Ceil(float64(total) / float64(r.Limit))),
			TotalItems: total,
			ItemsPerPage: r.Limit,
		},
	}, nil
}