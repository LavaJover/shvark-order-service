package usecase

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
)

type DefaultBankDetailUsecase struct {
	bankDetailRepo domain.BankDetailRepository
}

func NewDefaultBankDetailUsecase(bankDetailRepo domain.BankDetailRepository) *DefaultBankDetailUsecase {
	return &DefaultBankDetailUsecase{bankDetailRepo: bankDetailRepo}
}

func (uc *DefaultBankDetailUsecase) CreateBankDetail(bankDetail *domain.BankDetail) (string, error) {
	return uc.bankDetailRepo.CreateBankDetail(bankDetail)
}

func (uc *DefaultBankDetailUsecase) UpdateBankDetail(bankDetail *domain.BankDetail) error {
	return uc.bankDetailRepo.UpdateBankDetail(bankDetail)
}

func (uc *DefaultBankDetailUsecase) DeleteBankDetail(bankDetailID string) error {
	return uc.bankDetailRepo.DeleteBankDetail(bankDetailID)
}

func (uc *DefaultBankDetailUsecase) GetBankDetailByID(bankDetailID string) (*domain.BankDetail, error) {
	return uc.bankDetailRepo.GetBankDetailByID(bankDetailID)
}

func (uc *DefaultBankDetailUsecase) GetBankDetailsByTraderID(traderID string, page, limit int, sortBy, sortOrder string) ([]*domain.BankDetail, int64, error) {
	return uc.bankDetailRepo.GetBankDetailsByTraderID(
		traderID,
		page,
		limit,
		sortBy,
		sortOrder,
	)
}