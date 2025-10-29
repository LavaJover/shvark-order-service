package grpcapi

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/usdt"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	disputedto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/dispute"
	orderdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/order"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderHandler struct {
	uc usecase.OrderUsecase
	disputeUc usecase.DisputeUsecase
	bankDetailUc usecase.BankDetailUsecase
	orderpb.UnimplementedOrderServiceServer
}

func NewOrderHandler(
	uc usecase.OrderUsecase,
	disputeUc usecase.DisputeUsecase,
	bankDetailUc usecase.BankDetailUsecase,
	) *OrderHandler {
	return &OrderHandler{
		uc: uc,
		disputeUc: disputeUc,
		bankDetailUc: bankDetailUc,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, r *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	amountCrypto := r.AmountFiat / usdt.UsdtRubRates

	createOrderInput := orderdto.CreateOrderInput{
		MerchantParams: orderdto.MerchantParams{
			MerchantID: r.MerchantId,
			MerchantOrderID: r.MerchantOrderId,
			ClientID: r.ClientId,
		},
		PaymentSearchParams: orderdto.PaymentSearchParams{
			AmountFiat: r.AmountFiat,
			AmountCrypto: amountCrypto,
			Currency: r.Currency,
			CryptoRate: usdt.UsdtRubRates,
			PaymentSystem: r.PaymentSystem,
			BankInfo: orderdto.BankInfo{
				BankCode: r.BankCode,
				NspkCode: r.NspkCode,
			},
		},
		AdvancedParams: orderdto.AdvancedParams{
			Shuffle: r.Shuffle,
			CallbackUrl: r.CallbackUrl,
		},
		Type: "DEPOSIT",
		ExpiresAt: r.ExpiresAt.AsTime(),
	}
	
	createOrderOutput, err := h.uc.CreateOrder(&createOrderInput)
	if err != nil {
		if createOrderInput.AdvancedParams.CallbackUrl != ""{
			notifier.SendCallback(
				createOrderInput.AdvancedParams.CallbackUrl,
				createOrderInput.MerchantParams.MerchantOrderID,
				string(domain.StatusFailed),
				0, 0, 0,
			)
		}
		return nil, err
	}

	return &orderpb.CreateOrderResponse{
		Order: &orderpb.Order{
			OrderId: createOrderOutput.Order.ID,
			Status: string(createOrderOutput.Order.Status),
			Type: createOrderOutput.Order.Type,
			BankDetail: &orderpb.BankDetail{
				BankDetailId: createOrderOutput.BankDetail.ID,
				TraderId: createOrderOutput.BankDetail.TraderInfo.TraderID,
				Currency: createOrderOutput.Order.AmountInfo.Currency,
				Country: createOrderOutput.BankDetail.Country, 
				MinAmount: float64(createOrderOutput.BankDetail.MinOrderAmount),
				MaxAmount: float64(createOrderOutput.BankDetail.MaxOrderAmount),
				BankName: createOrderOutput.BankDetail.BankName,
				PaymentSystem: createOrderOutput.BankDetail.PaymentSystem,
				Owner: createOrderOutput.BankDetail.PaymentDetails.Owner,
				CardNumber: createOrderOutput.BankDetail.PaymentDetails.CardNumber,
				Phone: createOrderOutput.BankDetail.PaymentDetails.Phone,
				DeviceId: createOrderOutput.BankDetail.DeviceInfo.DeviceID,
				InflowCurrency: createOrderOutput.BankDetail.InflowCurrency,
				BankCode: createOrderOutput.BankDetail.PaymentDetails.BankCode,
				NspkCode: createOrderOutput.BankDetail.PaymentDetails.NspkCode,
			},
			AmountFiat: float64(createOrderOutput.Order.AmountInfo.AmountFiat),
			AmountCrypto: createOrderOutput.Order.AmountInfo.AmountCrypto,
			ExpiresAt: timestamppb.New(createOrderOutput.Order.ExpiresAt),
			Shuffle: createOrderOutput.Order.Shuffle,
			MerchantOrderId: createOrderOutput.Order.MerchantInfo.MerchantOrderID,
			ClientId: createOrderOutput.Order.MerchantInfo.ClientID,
			CallbackUrl: createOrderOutput.Order.CallbackUrl,
			TraderRewardPercent: createOrderOutput.Order.TraderReward,
			CreatedAt: timestamppb.New(createOrderOutput.Order.CreatedAt),
			UpdatedAt: timestamppb.New(createOrderOutput.Order.UpdatedAt),
			Recalculated: createOrderOutput.Order.Recalculated,
			CryptoRubRate: createOrderOutput.Order.AmountInfo.CryptoRate,
		},
	}, nil
}

func (h *OrderHandler) ApproveOrder(ctx context.Context, r *orderpb.ApproveOrderRequest) (*orderpb.ApproveOrderResponse, error) {
	orderID := r.OrderId
	if err := h.uc.ApproveOrder(orderID); err != nil {
		fmt.Println("Ошибка подтверждения сделки: ", err.Error())
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
	order, err := h.uc.GetOrderByID(orderID)
	if err != nil {
		return nil, err
	}

	return &orderpb.GetOrderByIDResponse{
		Order: &orderpb.Order{
			OrderId: orderID,
			Status: string(order.Status),
			Type: order.Type,
			BankDetail: &orderpb.BankDetail{
				BankDetailId: order.BankDetailID,
				TraderId: order.RequisiteDetails.TraderID,
				Currency: order.AmountInfo.Currency,
				Country: "unknown",
				BankName: order.RequisiteDetails.BankName,
				PaymentSystem: order.RequisiteDetails.PaymentSystem,
				Owner: order.RequisiteDetails.Owner,
				CardNumber: order.RequisiteDetails.CardNumber,
				Phone: order.RequisiteDetails.Phone,
				BankCode: order.RequisiteDetails.BankCode,
				NspkCode: order.RequisiteDetails.NspkCode,
				InflowCurrency: order.AmountInfo.Currency,
			},
			AmountFiat: order.AmountInfo.AmountFiat,
			AmountCrypto: order.AmountInfo.AmountCrypto,
			ExpiresAt: timestamppb.New(order.ExpiresAt),
			MerchantOrderId: order.MerchantInfo.MerchantOrderID,
			Shuffle: order.Shuffle,
			ClientId: order.MerchantInfo.ClientID,
			CallbackUrl: order.CallbackUrl,
			TraderRewardPercent: order.TraderReward,
			CreatedAt: timestamppb.New(order.CreatedAt),
			UpdatedAt: timestamppb.New(order.UpdatedAt),
			Recalculated: order.Recalculated,
			CryptoRubRate: order.AmountInfo.CryptoRate,
			MerchantId: order.MerchantInfo.MerchantID,
			Metrics: &orderpb.OrderMetrics{
				CompletedAt: timestamppb.New(order.Metrics.CompletedAt),
				CancelledAd: timestamppb.New(order.Metrics.CanceledAt),
				AutomaticCompleted: order.Metrics.AutomaticCompleted,
				ManuallyCompleted: order.Metrics.ManuallyCompleted,
			},
		},
	}, nil
}

func (h *OrderHandler) GetOrderByMerchantOrderID(ctx context.Context, r *orderpb.GetOrderByMerchantOrderIDRequest) (*orderpb.GetOrderByMerchantOrderIDResponse, error) {
	merchantOrderID := r.MerchantOrderId
	order, err := h.uc.GetOrderByMerchantOrderID(merchantOrderID)
	if err != nil {
		return nil, err
	}

	return &orderpb.GetOrderByMerchantOrderIDResponse{
		Order: &orderpb.Order{
			OrderId: order.ID,
			Status: string(order.Status),
			Type: order.Type,
			BankDetail: &orderpb.BankDetail{
				BankDetailId: order.BankDetailID,
				TraderId: order.RequisiteDetails.TraderID,
				Currency: order.AmountInfo.Currency,
				Country: "unknown",
				BankName: order.RequisiteDetails.BankName,
				PaymentSystem: order.RequisiteDetails.PaymentSystem,
				Owner: order.RequisiteDetails.Owner,
				CardNumber: order.RequisiteDetails.CardNumber,
				Phone: order.RequisiteDetails.Phone,
				BankCode: order.RequisiteDetails.BankCode,
				NspkCode: order.RequisiteDetails.NspkCode,
				InflowCurrency: order.AmountInfo.Currency,
			},
			AmountFiat: order.AmountInfo.AmountFiat,
			AmountCrypto: order.AmountInfo.AmountCrypto,
			ExpiresAt: timestamppb.New(order.ExpiresAt),
			MerchantOrderId: order.MerchantInfo.MerchantOrderID,
			Shuffle: order.Shuffle,
			ClientId: order.MerchantInfo.ClientID,
			CallbackUrl: order.CallbackUrl,
			TraderRewardPercent: order.TraderReward,
			CreatedAt: timestamppb.New(order.CreatedAt),
			UpdatedAt: timestamppb.New(order.UpdatedAt),
			Recalculated: order.Recalculated,
			CryptoRubRate: order.AmountInfo.CryptoRate,
			MerchantId: order.MerchantInfo.MerchantID,
			Metrics: &orderpb.OrderMetrics{
				CompletedAt: timestamppb.New(order.Metrics.CompletedAt),
				CancelledAd: timestamppb.New(order.Metrics.CanceledAt),
				AutomaticCompleted: order.Metrics.AutomaticCompleted,
				ManuallyCompleted: order.Metrics.ManuallyCompleted,
			},
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
		OrderID: r.Filters.OrderId,
		MerchantOrderID: r.Filters.MerchantOrderId,
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
			OrderId: order.Order.ID,
			Status: string(order.Order.Status),
			Type: order.Order.Type,
			BankDetail: &orderpb.BankDetail{
				BankDetailId: order.BankDetail.ID,
				TraderId: order.BankDetail.TraderID,
				Currency: order.BankDetail.Currency,
				Country: order.BankDetail.Country,
				MinAmount: float64(order.BankDetail.MinOrderAmount),
				MaxAmount: float64(order.BankDetail.MaxOrderAmount),
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
			AmountFiat: float64(order.Order.AmountInfo.AmountFiat),
			AmountCrypto: float64(order.Order.AmountInfo.AmountCrypto),
			ExpiresAt: timestamppb.New(order.Order.ExpiresAt),
			MerchantOrderId: order.Order.MerchantInfo.MerchantOrderID,
			Shuffle: order.Order.Shuffle,
			ClientId: order.Order.MerchantInfo.ClientID,
			CallbackUrl: order.Order.CallbackUrl,
			TraderRewardPercent: order.Order.TraderReward,
			CreatedAt: timestamppb.New(order.Order.CreatedAt),
			UpdatedAt: timestamppb.New(order.Order.UpdatedAt),
			Recalculated: order.Order.Recalculated,
			CryptoRubRate: order.Order.AmountInfo.CryptoRate,
			MerchantId: order.Order.MerchantInfo.MerchantID,
			Metrics: &orderpb.OrderMetrics{
				CompletedAt: timestamppb.New(order.Order.Metrics.CompletedAt),
				CancelledAd: timestamppb.New(order.Order.Metrics.CanceledAt),
				AutomaticCompleted: order.Order.Metrics.AutomaticCompleted,
				ManuallyCompleted: order.Order.Metrics.ManuallyCompleted,
			},
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
	createDisputeInput := disputedto.CreateDisputeInput{
		OrderID: r.OrderId,
		ProofUrl: r.ProofUrl,
		Ttl: r.Ttl.AsDuration(),
		DisputeAmountFiat: r.DisputeAmountFiat,
		DisputeAmountCrypto: r.DisputeAmountFiat / usdt.UsdtRubRates,
		DisputeCryptoRate: usdt.UsdtRubRates,
		Reason: r.DisputeReason,
	}
	if err := h.disputeUc.CreateDispute(&createDisputeInput); err != nil {
		return nil, err
	}
	return &orderpb.CreateOrderDisputeResponse{
		DisputeId: "",
	}, nil
}

func (h *OrderHandler) AcceptOrderDispute(ctx context.Context, r *orderpb.AcceptOrderDisputeRequest) (*orderpb.AcceptOrderDisputeResponse, error) {
	disputeID := r.DisputeId
	if err := h.disputeUc.AcceptDispute(disputeID); err != nil {
		slog.Error("failed to accept dispute", "error", err.Error())
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
	input := &disputedto.GetOrderDisputesInput{
		Page: r.Page,
		Limit: r.Limit,
		Status: r.Status,
		TraderID: r.TraderId,
		DisputeID: r.DisputeId,
		MerchantID: r.MerchantId,
		OrderID: r.OrderId,
	}
	output, err := h.disputeUc.GetOrderDisputes(input)
	if err != nil {
		return nil, err
	}

	disputesResp := make([]*orderpb.OrderDispute, len(output.Disputes))
	for i, dispute := range output.Disputes {
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
				AmountFiat: order.AmountInfo.AmountFiat,
				AmountCrypto: order.AmountInfo.AmountCrypto,
				ExpiresAt: timestamppb.New(order.ExpiresAt),
				MerchantOrderId: order.MerchantInfo.MerchantOrderID,
				TraderRewardPercent: order.TraderReward,
				CreatedAt: timestamppb.New(order.CreatedAt),
				UpdatedAt: timestamppb.New(order.UpdatedAt),
				CryptoRubRate: order.AmountInfo.CryptoRate,
				Type: order.Type,
				BankDetail: &orderpb.BankDetail{
					BankDetailId: order.BankDetailID,
					TraderId: order.RequisiteDetails.TraderID,
					Currency: order.AmountInfo.Currency,
					Country: "unknown",
					BankName: order.RequisiteDetails.BankName,
					PaymentSystem: order.RequisiteDetails.PaymentSystem,
					Owner: order.RequisiteDetails.Owner,
					CardNumber: order.RequisiteDetails.CardNumber,
					Phone: order.RequisiteDetails.Phone,
					BankCode: order.RequisiteDetails.BankCode,
					NspkCode: order.RequisiteDetails.NspkCode,
					InflowCurrency: order.AmountInfo.Currency,
				},
				MerchantId: order.MerchantInfo.MerchantID,
				Shuffle: order.Shuffle,
				ClientId: order.MerchantInfo.ClientID,
				CallbackUrl: order.CallbackUrl,
				Recalculated: order.Recalculated,
				Metrics: &orderpb.OrderMetrics{
					CompletedAt: timestamppb.New(order.Metrics.CompletedAt),
					CancelledAd: timestamppb.New(order.Metrics.CanceledAt),
					AutomaticCompleted: order.Metrics.AutomaticCompleted,
					ManuallyCompleted: order.Metrics.ManuallyCompleted,
				},
			},
		}
	}

	return &orderpb.GetOrderDisputesResponse{
		Disputes: disputesResp,
		Pagination: &orderpb.Pagination{
			CurrentPage: int64(output.Pagination.CurrentPage),
			TotalPages:  int64(output.Pagination.TotalPages),
			TotalItems: int64(output.Pagination.TotalItems),
			ItemsPerPage: int64(output.Pagination.ItemsPerPage),
		},
	}, nil
}

func (h *OrderHandler) GetOrderStatistics(ctx context.Context, r *orderpb.GetOrderStatisticsRequest) (*orderpb.GetOrderStatisticsResponse, error) {
	stats, err := h.uc.GetOrderStatistics(
		r.TraderId,
		r.DateFrom.AsTime(),
		r.DateTo.AsTime(),
	)

	if err != nil {
		return nil, err
	}

	return &orderpb.GetOrderStatisticsResponse{
		TotalOrders: stats.TotalOrders,
		SucceedOrders: stats.SucceedOrders,
		CanceledOrders: stats.CanceledOrders,
		ProcessedAmountFiat: float32(stats.ProcessedAmountFiat),
		ProcessedAmountCrypto: float32(stats.ProcessedAmountCrypto),
		CanceledAmountFiat: float32(stats.CanceledAmountFiat),
		CanceledAmountCrypto: float32(stats.CanceledAmountCrypto),
		IncomeCrypto: float32(stats.IncomeCrypto),
	}, nil
}

func (h *OrderHandler) GetOrders(ctx context.Context, r *orderpb.GetOrdersRequest) (*orderpb.GetOrdersResponse, error) {
    // Обработка параметра сортировки
    sortField := ""
    if r.Sort != nil {
        sortField = *r.Sort
    }

    page, size, merchantID := r.Page, r.Size, r.MerchantId
    if size <= 0 {
        size = 10
    }

    filter := domain.Filter{
        DealID:     r.DealId,
        Type:       r.Type,
        Status:     r.Status,
        AmountMin:  r.AmountMin,
        AmountMax:  r.AmountMax,
        MerchantID: merchantID,
    }

    if r.TimeOpeningStart != nil {
        t := r.TimeOpeningStart.AsTime()
        filter.TimeOpeningStart = &t
    }
    if r.TimeOpeningEnd != nil {
        t := r.TimeOpeningEnd.AsTime()
        filter.TimeOpeningEnd = &t
    }

    orders, total, err := h.uc.GetOrders(
        filter,
        sortField, // передаем строку, а не указатель
        int(page),
        int(size),
    )
    if err != nil {
        return nil, err
    }

    // ИСПРАВЛЕНО: создаем срез с 0 длиной и нужной емкостью
    content := make([]*orderpb.OrderResponse, 0, len(orders))
    for _, o := range orders {
        // ИСПРАВЛЕНО: используем реальные данные из модели
		// bankDetailID := o.BankDetailID
		// bankDetail, err := h.bankDetailUc.GetBankDetailByID(bankDetailID)
		// if err != nil {
		// 	return nil, err
		// }
        response := &orderpb.OrderResponse{
            Id:           o.MerchantInfo.MerchantOrderID,
            TimeOpening:  timestamppb.New(o.CreatedAt),
            TimeExpires:  timestamppb.New(o.ExpiresAt),
            TimeComplete: timestamppb.New(o.UpdatedAt),
            StoreName:    "UNKNOWN", // TODO: заменить на реальное значение
            Type:         o.Type,    // ИСПРАВЛЕНО: берем из модели
            Status:       string(o.Status),
            CurrencyRate: o.AmountInfo.CryptoRate,
            SumInvoice: &orderpb.Amount{
                Amount:   o.AmountInfo.AmountFiat,
                Currency: o.AmountInfo.Currency,
            },
            SumDeal: &orderpb.Amount{
                Amount:   o.AmountInfo.AmountCrypto - o.AmountInfo.AmountCrypto * o.PlatformFee, // ИСПРАВЛЕНО: используем AmountCrypto
                Currency: "USDT",
            },
            Requisites: &orderpb.Requisites{
                Issuer:      o.RequisiteDetails.BankCode,
                HolderName:  o.RequisiteDetails.Owner,
                PhoneNumber: o.RequisiteDetails.Phone,
				CardNumber: o.RequisiteDetails.CardNumber,
            },
            Email: "email", // TODO: заменить на реальное значение
        }
		if o.Status != domain.StatusCompleted {
			response.TimeComplete = nil
		}
        content = append(content, response)
    }

    // ИСПРАВЛЕНО: корректное вычисление пагинации
    totalPages := 0
    if size > 0 && total > 0 {
        totalPages = (int(total) + int(size) - 1) / int(size)
    }

    offset := page * size
    last := int(page) >= totalPages-1

    return &orderpb.GetOrdersResponse{
        Content: content,
        Pageable: &orderpb.Pageable{
            Sort: &orderpb.Sort{
                Unsorted: false,
                Sorted:   true,
                Empty:    false,
            },
            PageNumber: page,
            PageSize:   size,
            Offset:     offset,
            Paged:      true,
            Unpaged:    false,
        },
        TotalElements:    int32(total),
        TotalPages:       int32(totalPages),
        Last:             last,
        First:            page == 0,
        NumberOfElements: int32(len(content)),
        Size:             size,
        Number:           page,
        Sort: &orderpb.Sort{
            Unsorted: false,
            Sorted:   true,
            Empty:    false,
        },
        Empty: len(content) == 0,
    }, nil
}

func (h *OrderHandler) GetAllOrders(
    ctx context.Context, 
    r *orderpb.GetAllOrdersRequest,
) (*orderpb.GetAllOrdersResponse, error) {
    // Преобразуем protobuf в DTO
    input := &orderdto.GetAllOrdersInput{
        TraderID:        r.GetTraderId(),
        MerchantID:      r.GetMerchantId(),
        OrderID:         r.GetOrderId(),
        MerchantOrderID: r.GetMerchantOrderId(),
        Status:          r.GetStatus(),
        BankCode:        r.GetBankCode(),
        AmountFiatMin:   r.GetAmountMin(),
        AmountFiatMax:   r.GetAmountMax(),
        Type:            r.GetType(),
        DeviceID:        r.GetDeviceId(),
        Page:            r.GetPage(),
        Limit:           r.GetLimit(),
        Sort:            r.GetSort(),
		PaymentSystem: 	 r.GetPaymentSystem(),
    }

    // Обрабатываем временные метки
    if r.TimeOpeningStart != nil {
        input.TimeOpeningStart = r.TimeOpeningStart.AsTime()
    }
    if r.TimeOpeningEnd != nil {
        input.TimeOpeningEnd = r.TimeOpeningEnd.AsTime()
    }

    // Вызываем usecase
    output, err := h.uc.GetAllOrders(input)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    // Преобразуем результат в protobuf
    res := &orderpb.GetAllOrdersResponse{
        Orders: make([]*orderpb.Order, len(output.Orders)),
        Pagination: &orderpb.Pagination{
            CurrentPage:  int64(output.Pagination.CurrentPage),
            TotalPages:   int64(output.Pagination.TotalPages),
            TotalItems:   int64(output.Pagination.TotalItems),
            ItemsPerPage: int64(output.Pagination.ItemsPerPage),
        },
    }

    for i, order := range output.Orders {
        // Здесь используем ваш маппинг в protobuf
		// bankDetail, err := h.bankDetailUc.GetBankDetailByID(order.BankDetailID)
		// if err != nil {
		// 	return nil, err
		// }
        res.Orders[i] = &orderpb.Order{
			OrderId: order.ID,
			Status: string(order.Status),
			BankDetail: &orderpb.BankDetail{
				BankDetailId: order.BankDetailID,
				TraderId: order.RequisiteDetails.TraderID,
				Currency: order.AmountInfo.Currency,
				Country: "unknown",
				BankName: order.RequisiteDetails.BankName,
				PaymentSystem: order.RequisiteDetails.PaymentSystem,
				CardNumber: order.RequisiteDetails.CardNumber,
				Phone: order.RequisiteDetails.Phone,
				Owner: order.RequisiteDetails.Owner,
				DeviceId: order.RequisiteDetails.DeviceID,
				InflowCurrency: order.AmountInfo.Currency,
				BankCode: order.RequisiteDetails.BankCode,
				NspkCode: order.RequisiteDetails.NspkCode,
			},
			AmountFiat: order.AmountInfo.AmountFiat,
			AmountCrypto: order.AmountInfo.AmountCrypto,
			ExpiresAt: timestamppb.New(order.ExpiresAt),
			MerchantOrderId: order.MerchantInfo.MerchantOrderID,
			Shuffle: order.Shuffle,
			ClientId: order.MerchantInfo.ClientID,
			CallbackUrl: order.CallbackUrl,
			TraderRewardPercent: order.TraderReward,
			CreatedAt: timestamppb.New(order.CreatedAt),
			UpdatedAt: timestamppb.New(order.UpdatedAt),
			Recalculated: order.Recalculated,
			CryptoRubRate: order.AmountInfo.CryptoRate,
			MerchantId: order.MerchantInfo.MerchantID,
			Type: order.Type,
			Metrics: &orderpb.OrderMetrics{
				CompletedAt: timestamppb.New(order.Metrics.CompletedAt),
				CancelledAd: timestamppb.New(order.Metrics.CanceledAt),
				AutomaticCompleted: order.Metrics.AutomaticCompleted,
				ManuallyCompleted: order.Metrics.ManuallyCompleted,
			},
		}
    }

    return res, nil
}

// ProcessAutomaticPayment - обработчик gRPC для автоматического закрытия сделок
func (h *OrderHandler) ProcessAutomaticPayment(
	ctx context.Context, 
	req *orderpb.ProcessAutomaticPaymentRequest,
) (*orderpb.ProcessAutomaticPaymentResponse, error) {
	
	// Валидация входящего запроса
	if err := h.validateAutomaticPaymentRequest(req); err != nil {
		slog.Error("Invalid automatic payment request", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	// Преобразование gRPC запроса в доменную модель
	paymentReq := &usecase.AutomaticPaymentRequest{
		Group:         req.Group,
		Amount:        req.Amount,
		PaymentSystem: req.PaymentSystem,
		Direction: 	   req.Direction,
		Methods:       req.Methods,
		ReceivedAt:    req.ReceivedAt,
		Text:          req.Text,
		Metadata:      req.Metadata,
	}

	// Вызов usecase слоя
	result, err := h.uc.ProcessAutomaticPayment(ctx, paymentReq)
	if err != nil {
		slog.Error("Failed to process automatic payment", "error", err)
		return nil, status.Errorf(codes.Internal, "processing failed: %v", err)
	}

	// Преобразование результата в gRPC ответ
	return h.buildResponse(result), nil
}

func (h *OrderHandler) validateAutomaticPaymentRequest(req *orderpb.ProcessAutomaticPaymentRequest) error {
	if req.Group == "" {
		return status.Error(codes.InvalidArgument, "group is required")
	}
	if req.Amount <= 0 {
		return status.Error(codes.InvalidArgument, "amount must be positive")
	}
	if req.PaymentSystem == "" {
		return status.Error(codes.InvalidArgument, "payment_system is required")
	}
	if req.Direction != "in" {
		return status.Error(codes.InvalidArgument, "direction must be 'in'")
	}
	if req.ReceivedAt == 0 {
		return status.Error(codes.InvalidArgument, "received_at is required")
	}
	return nil
}

func (h *OrderHandler) buildResponse(result *domain.AutomaticPaymentResult) *orderpb.ProcessAutomaticPaymentResponse {
	response := &orderpb.ProcessAutomaticPaymentResponse{
		Action:  result.Action,
		Message: result.Message,
		Success: len(result.Results) > 0,
	}

	// Если есть результаты обработки отдельных ордеров
	if len(result.Results) > 0 {
		response.OrderId = result.Results[0].OrderID // Первый обработанный ордер
	}

	return response
}