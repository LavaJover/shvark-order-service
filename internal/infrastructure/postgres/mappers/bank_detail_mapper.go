package mappers

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
)

func ToDomainBankDetail(model *models.BankDetailModel) *domain.BankDetail {
	return &domain.BankDetail{
		ID: model.ID,
		SearchParams: domain.SearchParams{
			MaxOrdersSimultaneosly: model.MaxOrdersSimultaneosly,
			MaxAmountDay: model.MaxAmountDay,
			MaxAmountMonth: model.MaxAmountMonth,
			MaxQuantityDay: model.MaxQuantityDay,
			MaxQuantityMonth: model.MaxQuantityMonth,
			MinOrderAmount: model.MinAmount,
			MaxOrderAmount: model.MaxAmount,
			Delay: model.Delay,
			Enabled: model.Enabled,
		},
		DeviceInfo: domain.DeviceInfo{
			DeviceID: model.DeviceID,
		},
		TraderInfo: domain.TraderInfo{
			TraderID: model.TraderID,
		},
		PaymentDetails: domain.PaymentDetails{
			Phone: model.Phone,
			CardNumber: model.CardNumber,
			Owner: model.Owner,
			PaymentSystem: model.PaymentSystem,
			BankInfo: domain.BankInfo{
				BankCode: model.BankCode,
				BankName: model.BankName,
				NspkCode: model.NspkCode,
			},
		},
		Country: model.Country,
		Currency: model.Currency,
		InflowCurrency: model.InflowCurrency,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func ToGORMBankDetail(bankDetail *domain.BankDetail) *models.BankDetailModel {
	return &models.BankDetailModel{
		ID: bankDetail.ID,
	}
}