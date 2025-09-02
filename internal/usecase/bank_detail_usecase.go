package usecase

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	bankdetaildto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/bank_detail"
	"github.com/google/uuid"
)

type BankDetailUsecase interface {
	CreateBankDetail(input *bankdetaildto.CreateBankDetailInput) error
	UpdateBankDetail(input *bankdetaildto.UpdateBankDetailInput) error
	DeleteBankDetail(bankDetailID string) error
	GetBankDetailByID(bankDetailID string) (*domain.BankDetail, error)
	GetBankDetailsByTraderID(
		traderID string,
		page, limit int,
		sortBy, sortOrder string,
	) ([]*domain.BankDetail, int64, error)
	FindSuitableBankDetails(input *bankdetaildto.FindSuitableBankDetailsInput) ([]*domain.BankDetail, error)
	GetBankDetailsStatsByTraderID(traderID string) ([]*domain.BankDetailStat, error)
	GetBankDetails(input *bankdetaildto.GetBankDetailsInput) (*bankdetaildto.GetBankDetailsOutput, error)
}

type DefaultBankDetailUsecase struct {
	bankDetailRepo domain.BankDetailRepository
}

func NewDefaultBankDetailUsecase(bankDetailRepo domain.BankDetailRepository) *DefaultBankDetailUsecase {
	return &DefaultBankDetailUsecase{bankDetailRepo: bankDetailRepo}
}

func (uc *DefaultBankDetailUsecase) CreateBankDetail(input *bankdetaildto.CreateBankDetailInput) error {
	return uc.bankDetailRepo.CreateBankDetail(
		&domain.BankDetail{
			ID: uuid.New().String(),
			SearchParams: domain.SearchParams{
				MaxOrdersSimultaneosly: input.MaxOrdersSimultaneosly,
				MaxAmountDay: input.MaxAmountDay,
				MaxAmountMonth: input.MaxAmountMonth,
				MaxQuantityDay: input.MaxQuantityDay,
				MaxQuantityMonth: input.MaxQuantityMonth,
				MinOrderAmount: input.MinOrderAmount,
				MaxOrderAmount: input.MaxOrderAmount,
				Delay: input.Delay,
				Enabled: input.Enabled,
			},
			DeviceInfo: domain.DeviceInfo{
				DeviceID: input.DeviceID,
			},
			TraderInfo: domain.TraderInfo{
				TraderID: input.TraderID,
			},
			PaymentDetails: domain.PaymentDetails{
				Phone: input.Phone,
				CardNumber: input.CardNumber,
				Owner: input.Owner,
				PaymentSystem: input.PaymentSystem,
				BankInfo: domain.BankInfo{
					BankCode: input.BankCode,
					BankName: input.BankName,
					NspkCode: input.NspkCode,
				},
			},
			Country: input.Country,
			Currency: input.Currency,
			InflowCurrency: input.InflowCurrency,
		},
	)
}

func (uc *DefaultBankDetailUsecase) UpdateBankDetail(input *bankdetaildto.UpdateBankDetailInput) error {
	return uc.bankDetailRepo.UpdateBankDetail(
		&domain.BankDetail{
			ID: input.ID,
			SearchParams: domain.SearchParams{
				MaxOrdersSimultaneosly: input.MaxOrdersSimultaneosly,
				MaxAmountDay: input.MaxAmountDay,
				MaxAmountMonth: input.MaxAmountMonth,
				MaxQuantityDay: input.MaxQuantityDay,
				MaxQuantityMonth: input.MaxQuantityMonth,
				MinOrderAmount: input.MinOrderAmount,
				MaxOrderAmount: input.MaxOrderAmount,
				Delay: input.Delay,
				Enabled: input.Enabled,
			},
			DeviceInfo: domain.DeviceInfo{
				DeviceID: input.DeviceID,
			},
			TraderInfo: domain.TraderInfo{
				TraderID: input.TraderID,
			},
			PaymentDetails: domain.PaymentDetails{
				Phone: input.Phone,
				CardNumber: input.CardNumber,
				Owner: input.Owner,
				PaymentSystem: input.PaymentSystem,
				BankInfo: domain.BankInfo{
					BankCode: input.BankCode,
					BankName: input.BankName,
					NspkCode: input.NspkCode,
				},
			},
			Country: input.Country,
			Currency: input.Currency,
			InflowCurrency: input.InflowCurrency,
		},
	)
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

func (uc *DefaultBankDetailUsecase) FindSuitableBankDetails(input *bankdetaildto.FindSuitableBankDetailsInput) ([]*domain.BankDetail, error) {
	return uc.bankDetailRepo.FindSuitableBankDetails(
		&domain.SuitablleBankDetailsQuery{
			AmountFiat: input.AmountFiat,
			BankCode: input.BankCode,
			NspkCode: input.NspkCode,
			PaymentSystem: input.PaymentSystem,
			Currency: input.Currency,
		},
	)
}

func (uc *DefaultBankDetailUsecase) GetBankDetailsStatsByTraderID(traderID string) ([]*domain.BankDetailStat, error) {
	return uc.bankDetailRepo.GetBankDetailsStatsByTraderID(traderID)
}

func (uc *DefaultBankDetailUsecase) GetBankDetails(input *bankdetaildto.GetBankDetailsInput) (*bankdetaildto.GetBankDetailsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 100 {
		input.Limit = 50
	}

	filter := domain.GetBankDetailsFilter{
		TraderID: input.TraderID,
		BankCode: input.BankCode,
		Enabled: input.Enabled,
		PaymentSystem: input.PaymentSystem,
		Page: input.Page,
		Limit: input.Limit,
	}

	bankDetails, total, err := uc.bankDetailRepo.GetBankDetails(filter)
	if err != nil {
		return nil, err
	}

	totalPages := total / int64(input.Limit)
	if total%int64(input.Limit) > 0 {
		totalPages++
	}

	return &bankdetaildto.GetBankDetailsOutput{
		BankDetails: bankDetails,
		Pagination: bankdetaildto.Pagination{
			CurrentPage: int32(input.Page),
			TotalPages: int32(totalPages),
			TotalItems: int32(total),
			ItemsPerPage: int32(input.Limit),
		},
	}, nil
}