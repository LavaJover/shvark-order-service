package grpcapi

import (
	"context"
	"github.com/LavaJover/shvark-order-service/internal/delivery/grpcapi/mappers"
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

func (h *UncreatedOrderHandler) GetUncreatedOrdersWithFilter(ctx context.Context, r *orderpb.GetUncreatedOrdersWithFilterRequest) (*orderpb.GetUncreatedOrdersWithFilterResponse, error) {
	domainFilter := mappers.ToDomainUncreatedOrdersFilter(r.GetFilters())
	output, err := h.uc.GetUncreatedLogsWithFilter(domainFilter, r.GetPage(), r.GetLimit(), r.GetSortBy(), r.GetSortOrder())

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &orderpb.GetUncreatedOrdersWithFilterResponse{
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

func (h *UncreatedOrderHandler) GetStatsForUncreatedOrdersWithFilter(ctx context.Context, r *orderpb.GetStatsForUncreatedOrdersWithFilterRequest) (*orderpb.GetStatsForUncreatedOrdersWithFilterResponse, error) {
	domainFilter := mappers.ToDomainUncreatedOrdersFilter(r.GetFilters())
	output, err := h.uc.GetStatsForUncreatedOrdersWithFilter(domainFilter, r.GetGroupByCriteria())

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &orderpb.GetStatsForUncreatedOrdersWithFilterResponse{
		Stats: make([]*orderpb.UncreatedOrdersStats, len(output)),
	}

	for i, uncreatedOrder := range output {
		res.Stats[i] = &orderpb.UncreatedOrdersStats{
			MerchantId:      uncreatedOrder.MerchantID,
			Currency:        uncreatedOrder.Currency,
			PaymentSystem:   uncreatedOrder.PaymentSystem,
			DateGroup:       uncreatedOrder.DateGroup,
			AmountRange:     uncreatedOrder.AmountRange,
			BankCode:        uncreatedOrder.BankCode,
			TotalCount:      uncreatedOrder.TotalCount,
			TotalAmountFiat: uncreatedOrder.TotalAmountFiat,
			AvgAmountFiat:   uncreatedOrder.AvgAmountFiat,
			MinAmountFiat:   uncreatedOrder.MinAmountFiat,
			MaxAmountFiat:   uncreatedOrder.MaxAmountFiat,
		}
	}
	return res, nil
}
