package grpcapi

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
)

type BankDetailHandler struct {
	bankDetailUsecase domain.BankDetailUsecase
	orderpb.UnimplementedBankDetailServiceServer
}

func NewBankDetailHandler(bankDetailUsecase domain.BankDetailUsecase) *BankDetailHandler {
	return &BankDetailHandler{bankDetailUsecase: bankDetailUsecase}
}

func (h *BankDetailHandler) CreateBankDetail(ctx context.Context, r *orderpb.CreateBankDetailRequest) (*orderpb.CreateBankDetailResponse, error) {

}

func (h *BankDetailHandler) UpdateBankDetail(ctx context.Context, r *orderpb.UpdateBankDetailRequest) (*orderpb.UpdateBankDetailResponse, error) {

}

func (h *BankDetailHandler) DeleteBankDetail(ctx context.Context, r *orderpb.DeleteBankDetailRequest) (*orderpb.DeleteBankDetailResponse, error) {

}

func (h *BankDetailHandler) GetBankDetailByID(ctx context.Context, r *orderpb.GetBankDetailByIDRequest) (*orderpb.GetBankDetailByIDResponse, error) {

}

func (h *BankDetailHandler) GetBankDetailsByTraderID(ctx context.Context, r *orderpb.GetBankDetailsByTraderIDRequest) (*orderpb.GetBankDetailsByTraderIDResponse, error) {
	
}