package postgres

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"gorm.io/gorm"
)

type DefaultBankDetailRepo struct {
	DB *gorm.DB
}

func NewDefaultBankDetailRepo(db *gorm.DB) *DefaultBankDetailRepo {
	return &DefaultBankDetailRepo{DB: db}
}

func (r *DefaultBankDetailRepo) SaveBankDetail(bankDetail *domain.BankDetail) error {
	bankDetailModel := BankDetailModel{
		ID: bankDetail.ID,
		TraderID: bankDetail.TraderID,
		Country: bankDetail.Country,
		Currency: bankDetail.Currency,
		MinAmount: bankDetail.MinAmount,
		MaxAmount: bankDetail.MaxAmount,
		BankName: bankDetail.BankName,
		PaymentSystem: bankDetail.PaymentSystem,
		Delay: bankDetail.Delay,
		Enabled: bankDetail.Enabled,
		CardNumber: bankDetail.CardNumber,
		Phone: bankDetail.Phone,
		Owner: bankDetail.Owner,
		MaxOrdersSimultaneosly: bankDetail.MaxOrdersSimultaneosly,
		MaxAmountDay: bankDetail.MaxAmountDay,
		MaxAmountMonth: bankDetail.MaxAmountMonth,
	}

	return r.DB.Save(&bankDetailModel).Error
}