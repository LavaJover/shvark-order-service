package grpcapi

import (
	"context"
	"math"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/protobuf/types/known/durationpb"
)

type BankDetailHandler struct {
	bankDetailUsecase domain.BankDetailUsecase
	orderpb.UnimplementedBankDetailServiceServer
}

func NewBankDetailHandler(bankDetailUsecase domain.BankDetailUsecase) *BankDetailHandler {
	return &BankDetailHandler{bankDetailUsecase: bankDetailUsecase}
}

func (h *BankDetailHandler) CreateBankDetail(ctx context.Context, r *orderpb.CreateBankDetailRequest) (*orderpb.CreateBankDetailResponse, error) {
	bankDetail := domain.BankDetail{
		TraderID: r.TraderId,
		Country: r.Country,
		Currency: r.Currency,
		InflowCurrency: r.InflowCurrency,
		MinAmount: float32(r.MinAmount),
		MaxAmount: float32(r.MaxAmount),
		BankName: r.BankName,
		PaymentSystem: r.PaymentSystem,
		Delay: r.Delay.AsDuration(),
		Enabled: r.Enabled,
		CardNumber: r.CardNumber,
		Phone: r.Phone,
		Owner: r.Owner,
		MaxOrdersSimultaneosly: r.MaxOrdersSimultaneosly,
		MaxAmountDay: int32(r.MaxAmountDay),
		MaxAmountMonth: int32(r.MaxAmountMonth),
		MaxQuantityDay: int32(r.MaxQuantityDay),
		MaxQuantityMonth: int32(r.MaxQuantityMonth),
		DeviceID: r.DeviceId,
	}

	bankDetailID, err := h.bankDetailUsecase.CreateBankDetail(&bankDetail)
	if err != nil {
		return nil, err
	}
	return &orderpb.CreateBankDetailResponse{
		BankDetailId: bankDetailID,
	}, nil
}

func (h *BankDetailHandler) UpdateBankDetail(ctx context.Context, r *orderpb.UpdateBankDetailRequest) (*orderpb.UpdateBankDetailResponse, error) {
	if r.BankDetail.Delay == nil {
		r.BankDetail.Delay = durationpb.New(0*time.Second)
	}
	bankDetail := domain.BankDetail{
		ID: r.BankDetail.BankDetailId,
		TraderID: r.BankDetail.TraderId,
		Country: r.BankDetail.Country,
		Currency: r.BankDetail.Currency,
		InflowCurrency: r.BankDetail.InflowCurrency,
		MinAmount: float32(r.BankDetail.MinAmount),
		MaxAmount: float32(r.BankDetail.MaxAmount),
		BankName: r.BankDetail.BankName,
		PaymentSystem: r.BankDetail.PaymentSystem,
		Delay: r.BankDetail.Delay.AsDuration(),
		Enabled: r.BankDetail.Enabled,
		CardNumber: r.BankDetail.CardNumber,
		Phone: r.BankDetail.Phone,
		Owner: r.BankDetail.Owner,
		MaxOrdersSimultaneosly: r.BankDetail.MaxOrdersSimultaneosly,
		MaxAmountDay: int32(r.BankDetail.MaxAmountDay),
		MaxAmountMonth: int32(r.BankDetail.MaxAmountMonth),
		MaxQuantityDay: int32(r.BankDetail.MaxQuantityDay),
		MaxQuantityMonth: int32(r.BankDetail.MaxQuantityMonth),
		DeviceID: r.BankDetail.DeviceId,
	}
	err := h.bankDetailUsecase.UpdateBankDetail(&bankDetail)
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
			MinAmount: float64(bankDetail.MinAmount),
			MaxAmount: float64(bankDetail.MaxAmount),
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
		},
	}, nil
}

func (h *BankDetailHandler) GetBankDetailsByTraderID(ctx context.Context, r *orderpb.GetBankDetailsByTraderIDRequest) (*orderpb.GetBankDetailsByTraderIDResponse, error) {
	traderID, page, limit, sortBy, sortOrder := r.TraderId, r.Page, r.Limit, r.SortBy, r.SortOrder
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
			MinAmount: float64(bankDetail.MinAmount),
			MaxAmount: float64(bankDetail.MaxAmount),
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