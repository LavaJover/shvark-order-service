package grpcapi

import (
	"context"
	"math"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/delivery/grpcapi/mappers"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	bankdetaildto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/bank_detail"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen/order"
	"google.golang.org/protobuf/types/known/durationpb"
)

type BankDetailHandler struct {
	bankDetailUsecase usecase.BankDetailUsecase
	orderpb.UnimplementedBankDetailServiceServer
}

func NewBankDetailHandler(bankDetailUsecase usecase.BankDetailUsecase) *BankDetailHandler {
	return &BankDetailHandler{bankDetailUsecase: bankDetailUsecase}
}

func (h *BankDetailHandler) CreateBankDetail(ctx context.Context, r *orderpb.CreateBankDetailRequest) (*orderpb.CreateBankDetailResponse, error) {
	createBankDetailInput := bankdetaildto.CreateBankDetailInput{
		SearchParams: bankdetaildto.SearchParams{
			MaxOrdersSimultaneosly: r.MaxOrdersSimultaneosly,
			MaxAmountDay: r.MaxAmountDay,
			MaxAmountMonth: r.MaxAmountMonth,
			MaxQuantityDay: int32(r.MaxQuantityDay),
			MaxQuantityMonth: int32(r.MaxQuantityMonth),
			MinOrderAmount: float32(r.MinAmount),
			MaxOrderAmount: float32(r.MaxAmount),
			Delay: r.Delay.AsDuration(),
			Enabled: r.Enabled,
		},
		DeviceInfo: bankdetaildto.DeviceInfo{
			DeviceID: r.DeviceId,
		},
		TraderInfo: bankdetaildto.TraderInfo{
			TraderID: r.TraderId,
		},
		PaymentDetails: bankdetaildto.PaymentDetails{
			Phone: r.Phone,
			CardNumber: r.CardNumber,
			Owner: r.Owner,
			PaymentSystem: r.PaymentSystem,
			BankInfo: bankdetaildto.BankInfo{
				BankCode: r.BankCode,
				BankName: r.BankName,
				NspkCode: r.NspkCode,
			},
		},
		Country: r.Country,
		Currency: r.Currency,
		InflowCurrency: r.InflowCurrency,
	}

	err := h.bankDetailUsecase.CreateBankDetail(&createBankDetailInput)
	if err != nil {
		return nil, err
	}
	return &orderpb.CreateBankDetailResponse{
		BankDetailId: "",
	}, nil
}

func (h *BankDetailHandler) UpdateBankDetail(ctx context.Context, r *orderpb.UpdateBankDetailRequest) (*orderpb.UpdateBankDetailResponse, error) {
	if r.BankDetail.Delay == nil {
		r.BankDetail.Delay = durationpb.New(0*time.Second)
	}
	updateBankDetailInput := bankdetaildto.UpdateBankDetailInput{
		ID: r.BankDetail.BankDetailId,
		SearchParams: bankdetaildto.SearchParams{
			MaxOrdersSimultaneosly: r.BankDetail.MaxOrdersSimultaneosly,
			MaxAmountDay: r.BankDetail.MaxAmountDay,
			MaxAmountMonth: r.BankDetail.MaxAmountMonth,
			MaxQuantityDay: int32(r.BankDetail.MaxQuantityDay),
			MaxQuantityMonth: int32(r.BankDetail.MaxQuantityMonth),
			MinOrderAmount: float32(r.BankDetail.MinAmount),
			MaxOrderAmount: float32(r.BankDetail.MaxAmount),
			Delay: r.BankDetail.Delay.AsDuration(),
			Enabled: r.BankDetail.Enabled,
		},
		DeviceInfo: bankdetaildto.DeviceInfo{
			DeviceID: r.BankDetail.DeviceId,
		},
		TraderInfo: bankdetaildto.TraderInfo{
			TraderID: r.BankDetail.TraderId,
		},
		PaymentDetails: bankdetaildto.PaymentDetails{
			Phone: r.BankDetail.Phone,
			CardNumber: r.BankDetail.CardNumber,
			Owner: r.BankDetail.Owner,
			PaymentSystem: r.BankDetail.PaymentSystem,
			BankInfo: bankdetaildto.BankInfo{
				BankCode: r.BankDetail.BankCode,
				BankName: r.BankDetail.BankName,
				NspkCode: r.BankDetail.NspkCode,
			},
		},
		Country: r.BankDetail.Country,
		Currency: r.BankDetail.Currency,
		InflowCurrency: r.BankDetail.InflowCurrency,
	}
	err := h.bankDetailUsecase.UpdateBankDetail(&updateBankDetailInput)
	if err != nil {
		return nil, err
	}

	return &orderpb.UpdateBankDetailResponse{}, nil
}

func (h *BankDetailHandler) DeleteBankDetail(ctx context.Context, r *orderpb.DeleteBankDetailRequest) (*orderpb.DeleteBankDetailResponse, error) {
	bankDetailID := r.BankDetailId
	err := h.bankDetailUsecase.DeleteBankDetail(bankDetailID)
	if err != nil {
		return nil, err 
	}

	return &orderpb.DeleteBankDetailResponse{}, nil
}

func (h *BankDetailHandler) GetBankDetailByID(ctx context.Context, r *orderpb.GetBankDetailByIDRequest) (*orderpb.GetBankDetailByIDResponse, error) {
	bankDetailID := r.BankDetailId
	bankDetail, err := h.bankDetailUsecase.GetBankDetailByID(bankDetailID)
	if err != nil {
		return nil, err
	}

	return &orderpb.GetBankDetailByIDResponse{
		BankDetail: &orderpb.BankDetail{
			BankDetailId: bankDetail.ID,
			TraderId: bankDetail.TraderID,
			Currency: bankDetail.Currency,
			Country: bankDetail.Country,
			MinAmount: float64(bankDetail.MinOrderAmount),
			MaxAmount: float64(bankDetail.MaxOrderAmount),
			BankName: bankDetail.BankName,
			PaymentSystem: bankDetail.PaymentSystem,
			Enabled: bankDetail.Enabled,
			Delay: durationpb.New(bankDetail.Delay),
			CardNumber: bankDetail.CardNumber,
			Phone: bankDetail.Phone,
			Owner: bankDetail.Owner,
			MaxOrdersSimultaneosly: bankDetail.MaxOrdersSimultaneosly,
			MaxAmountDay: float64(bankDetail.MaxAmountDay),
			MaxAmountMonth: float64(bankDetail.MaxAmountMonth),
			MaxQuantityDay: float64(bankDetail.MaxQuantityDay),
			MaxQuantityMonth: float64(bankDetail.MaxQuantityMonth),
			DeviceId: bankDetail.DeviceID,
			InflowCurrency: bankDetail.InflowCurrency,
			BankCode: bankDetail.BankCode,
			NspkCode: bankDetail.NspkCode,
		},
	}, nil
}

func (h *BankDetailHandler) GetBankDetailsByTraderID(ctx context.Context, r *orderpb.GetBankDetailsByTraderIDRequest) (*orderpb.GetBankDetailsByTraderIDResponse, error) {
	traderID, page, limit, sortBy, sortOrder := r.TraderId, r.Page, r.Limit, r.SortBy, r.SortOrder
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 1000
	}
	bankDetails, total, err := h.bankDetailUsecase.GetBankDetailsByTraderID(traderID, int(page), int(limit), sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	bankDetailsResp := make([]*orderpb.BankDetail, len(bankDetails))
	for i, bankDetail := range bankDetails{
		bankDetailsResp[i] = &orderpb.BankDetail{
			BankDetailId: bankDetail.ID,
			TraderId: bankDetail.TraderID,
			Currency: bankDetail.Currency,
			Country: bankDetail.Country,
			MinAmount: float64(bankDetail.MinOrderAmount),
			MaxAmount: float64(bankDetail.MaxOrderAmount),
			BankName: bankDetail.BankName,
			PaymentSystem: bankDetail.PaymentSystem,
			Enabled: bankDetail.Enabled,
			Delay: durationpb.New(bankDetail.Delay),
			CardNumber: bankDetail.CardNumber,
			Phone: bankDetail.Phone,
			Owner: bankDetail.Owner,
			MaxOrdersSimultaneosly: bankDetail.MaxOrdersSimultaneosly,
			MaxAmountDay: float64(bankDetail.MaxAmountDay),
			MaxAmountMonth: float64(bankDetail.MaxAmountMonth),
			MaxQuantityDay: float64(bankDetail.MaxQuantityDay),
			MaxQuantityMonth: float64(bankDetail.MaxQuantityMonth),
			DeviceId: bankDetail.DeviceID,
			InflowCurrency: bankDetail.InflowCurrency,
			BankCode: bankDetail.BankCode,
			NspkCode: bankDetail.NspkCode,
		}
	}

	return &orderpb.GetBankDetailsByTraderIDResponse{
		BankDetails: bankDetailsResp,
		Pagination: &orderpb.Pagination{
			CurrentPage: int64(page),
			TotalItems: total,
			TotalPages: int64(math.Ceil(float64(total) / float64(r.Limit))),
			ItemsPerPage: int64(limit),
		},
	}, nil

}

func (h *BankDetailHandler) GetBankDetailsStatsByTraderID(ctx context.Context, r *orderpb.GetBankDetailsStatsByTraderIDRequest) (*orderpb.GetBankDetailsStatsByTraderIDResponse, error) {
	traderID := r.TraderId
	response, err := h.bankDetailUsecase.GetBankDetailsStatsByTraderID(traderID)
	if err != nil {
		return nil, err
	}
	bankDetailsStats := make([]*orderpb.BankDetailStat, len(response))
	for i, bankDetailStat := range response {
		bankDetailsStats[i] = &orderpb.BankDetailStat{
			BankDetailId: bankDetailStat.BankDetailID,
			CurrentCountToday: int32(bankDetailStat.CurrentCountToday),
			CurrentCountMonth: int32(bankDetailStat.CurrentCountMonth),
			CurrentAmountToday: bankDetailStat.CurrentAmountToday,
			CurrentAmountMonth: bankDetailStat.CurrentAmountMonth,
		}
	}

	return &orderpb.GetBankDetailsStatsByTraderIDResponse{
		BankDetailStat: bankDetailsStats,
	}, nil
}

func (h *BankDetailHandler) GetBankDetails(ctx context.Context, r *orderpb.GetBankDetailsRequest) (*orderpb.GetBankDetailsResponse, error) {
	input := bankdetaildto.GetBankDetailsInput{
		TraderID: r.TraderId,
		PaymentSystem: r.PaymentSystem,
		BankCode: r.BankCode,
		Enabled: r.Enabled,
		Page: int(r.Page),
		Limit: int(r.Limit),
	}
	output, err := h.bankDetailUsecase.GetBankDetails(&input)
	if err != nil {
		return nil, err
	}

	response := &orderpb.GetBankDetailsResponse{
		BankDetails: make([]*orderpb.BankDetail, len(output.BankDetails)),
		Pagination: &orderpb.Pagination{
			CurrentPage: int64(output.Pagination.CurrentPage),
			TotalPages: int64(output.Pagination.TotalPages),
			TotalItems: int64(output.Pagination.TotalItems),
			ItemsPerPage: int64(output.Pagination.ItemsPerPage),
		},
	}

	for i, bankDetail := range output.BankDetails {
		response.BankDetails[i] = mappers.ToProtoBankDetail(bankDetail)
	}

	return response, nil
}