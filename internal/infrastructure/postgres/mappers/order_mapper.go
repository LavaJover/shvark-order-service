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
		Type: domain.OrderType(model.Type),
		Recalculated: model.Recalculated,
		Shuffle: model.Shuffle,
		TraderReward: model.TraderRewardPercent,
		PlatformFee: model.PlatformFee,
		CallbackUrl: model.CallbackURL,
		ExpiresAt: model.ExpiresAt,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,

		RequisiteDetails: domain.RequisiteDetails{
			TraderID: model.TraderID,
			CardNumber: model.CardNumber,
			Phone: model.Phone,
			Owner: model.Owner,
			PaymentSystem: model.PaymentSystem,
			BankName: model.BankName,
			BankCode: model.BankCode,
			NspkCode: model.NspkCode,
			DeviceID: model.DeviceID,
		},

		Metrics: domain.Metrics{
			AutomaticCompleted: model.AutomaticCompleted,
			ManuallyCompleted: model.ManuallyCompleted,
			CompletedAt: model.CompletedAt,
			CanceledAt: model.CanceledAt,
		},
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
		Type: string(order.Type),
		ExpiresAt: order.ExpiresAt,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,

		TraderID: order.RequisiteDetails.TraderID,
		CardNumber: order.RequisiteDetails.CardNumber,
		Phone: order.RequisiteDetails.Phone,
		Owner: order.RequisiteDetails.Owner,
		PaymentSystem: order.RequisiteDetails.PaymentSystem,
		BankName: order.RequisiteDetails.BankName,
		BankCode: order.RequisiteDetails.BankCode,
		NspkCode: order.RequisiteDetails.NspkCode,
		DeviceID: order.RequisiteDetails.DeviceID,
		AutomaticCompleted: order.Metrics.AutomaticCompleted,
		ManuallyCompleted: order.Metrics.ManuallyCompleted,
		CompletedAt: order.Metrics.CompletedAt,
		CanceledAt: order.Metrics.CanceledAt,
	}
}