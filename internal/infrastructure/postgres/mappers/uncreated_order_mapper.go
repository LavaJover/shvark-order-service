package mappers

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
)

func ToDomainUncreatedOrder(model *models.UncreatedOrderModel) *domain.UncreatedOrder {
	return &domain.UncreatedOrder{
		ID:              model.ID,
		MerchantID:      model.MerchantID,
		AmountFiat:      model.AmountFiat,
		AmountCrypto:    model.AmountCrypto,
		Currency:        model.Currency,
		ClientID:        model.ClientID,
		CreatedAt:       model.CreatedAt,
		MerchantOrderID: model.MerchantOrderID,
		PaymentSystem:   model.PaymentSystem,
		BankCode:        model.BankCode,
		ErrorMessage:    model.ErrorMessage,
	}
}

func ToGORMUncreatedOrder(uncreatedLog *domain.UncreatedOrder) *models.UncreatedOrderModel {
	return &models.UncreatedOrderModel{
		ID:              uncreatedLog.ID,
		MerchantID:      uncreatedLog.MerchantID,
		AmountFiat:      uncreatedLog.AmountFiat,
		AmountCrypto:    uncreatedLog.AmountCrypto,
		Currency:        uncreatedLog.Currency,
		ClientID:        uncreatedLog.ClientID,
		CreatedAt:       uncreatedLog.CreatedAt,
		MerchantOrderID: uncreatedLog.MerchantOrderID,
		PaymentSystem:   uncreatedLog.PaymentSystem,
		BankCode:        uncreatedLog.BankCode,
		ErrorMessage:    uncreatedLog.ErrorMessage,
	}
}
