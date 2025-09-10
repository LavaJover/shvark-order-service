package grpcapi

import (
	"context"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UncreatedOrderHandler struct {
	uc usecase.UncreatedOrderUsecase
	orderpb.UnimplementedUncreatedOrderServiceServer
}

func NewUncreatedOrderHandler(uc usecase.UncreatedOrderUsecase) *UncreatedOrderHandler {
	return &UncreatedOrderHandler{
		uc: uc,
	}
}

func (h *UncreatedOrderHandler) GetUncreatedOrdersWithFilters(ctx context.Context, r *orderpb.GetUncreatedOrdersWithFiltersRequest) (*orderpb.GetUncreatedOrdersWithFiltersResponse, error) {
	filter := r.GetFilters()
	var domainFilter *domain.UncreatedOrdersFilter = nil

	if filter != nil {
		domainFilter = &domain.UncreatedOrdersFilter{
			MerchantID:    r.Filters.MerchantId,
			MinAmountFiat: r.Filters.MinAmountFiat,
			MaxAmountFiat: r.Filters.MaxAmountFiat,
			Currency:      r.Filters.Currency,
			ClientID:      r.Filters.ClientId,
			PaymentSystem: r.Filters.PaymentSystem,
			BankCode:      r.Filters.BankCode,
		}

		if r.Filters.DateFrom != nil {
			dateFrom := r.Filters.DateFrom.AsTime()
			domainFilter.TimeOpeningStart = &dateFrom
		}

		if r.Filters.DateTo != nil {
			dateTo := r.Filters.DateTo.AsTime()
			domainFilter.TimeOpeningEnd = &dateTo
		}
	}

	output, err := h.uc.GetUncreatedLogsWithFilters(domainFilter, r.GetPage(), r.GetLimit(), r.GetSortBy(), r.GetSortOrder())

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &orderpb.GetUncreatedOrdersWithFiltersResponse{
		UncreatedOrders: make([]*orderpb.UncreatedOrder, len(output.UncreatedOrders)),
		Pagination: &orderpb.Pagination{
			CurrentPage:  int64(output.Pagination.CurrentPage),
			TotalPages:   int64(output.Pagination.TotalPages),
			TotalItems:   int64(output.Pagination.TotalItems),
			ItemsPerPage: int64(output.Pagination.ItemsPerPage),
		},
	}

	for i, uncreatedOrder := range output.UncreatedOrders {
		res.UncreatedOrders[i] = &orderpb.UncreatedOrder{
			OrderId:         uncreatedOrder.ID,
			MerchantId:      uncreatedOrder.MerchantID,
			AmountFiat:      uncreatedOrder.AmountFiat,
			AmountCrypto:    uncreatedOrder.AmountCrypto,
			Currency:        uncreatedOrder.Currency,
			ClientId:        uncreatedOrder.ClientID,
			CreatedAt:       timestamppb.New(uncreatedOrder.CreatedAt),
			MerchantOrderId: uncreatedOrder.MerchantOrderID,
			PaymentSystem:   uncreatedOrder.PaymentSystem,
			BankCode:        uncreatedOrder.BankCode,
			ErrorMessage:    uncreatedOrder.ErrorMessage,
		}
	}

	return res, nil
}
