package mappers

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
)

func ToDomainOrder(model *models.OrderModel) *domain.Order {
	return &domain.Order{
		ID: model.ID,
		Status: model.Status,
		MerchantInfo: domain.MerchantInfo{
			MerchantID: model.MerchantID,
			MerchantOrderID: model.MerchantOrderID,
			ClientID: model.ClientID,
		},
		AmountInfo: domain.AmountInfo{
			AmountFiat: model.AmountFiat,
			AmountCrypto: model.AmountCrypto,
			CryptoRate: model.CryptoRubRate,
			Currency: model.Currency,
		},
		BankDetailID: model.BankDetailsID,
		Type: model.Type,
		Recalculated: model.Recalculated,
		Shuffle: model.Shuffle,
		TraderReward: model.TraderRewardPercent,
		PlatformFee: model.PlatformFee,
		CallbackUrl: model.CallbackURL,
		ExpiresAt: model.ExpiresAt,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func ToGORMOrder(order *domain.Order) *models.OrderModel {
	return &models.OrderModel{
		ID: order.ID,
		MerchantID: order.MerchantInfo.MerchantID,
		AmountFiat: order.AmountInfo.AmountFiat,
		AmountCrypto: order.AmountInfo.AmountCrypto,
		Currency: order.AmountInfo.Currency,
		ClientID: order.MerchantInfo.ClientID,
		Status: order.Status,
		BankDetailsID: order.BankDetailID,
		MerchantOrderID: order.MerchantInfo.MerchantOrderID,
		Shuffle: order.Shuffle,
		CallbackURL: order.CallbackUrl,
		TraderRewardPercent: order.TraderReward,
		PlatformFee: order.PlatformFee,
		Recalculated: order.Recalculated,
		CryptoRubRate: order.AmountInfo.CryptoRate,
		Type: order.Type,
		ExpiresAt: order.ExpiresAt,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}