package usecase

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
)

// recordOrderCreatedMetrics - вызывается при создании заказа
func (uc *DefaultOrderUsecase) recordOrderCreatedMetrics(order *domain.Order, paymentSystem string) {
	if uc.Metrics == nil {
		return
	}

	orderType := string(order.Type)
	currency := order.AmountInfo.Currency
	amountFiat := order.AmountInfo.AmountFiat

	uc.Metrics.RecordOrderCreated(
		order.MerchantInfo.MerchantID,
		orderType,
		currency,
		paymentSystem,
		amountFiat,
	)

	// Записываем комиссии
	uc.Metrics.RecordPlatformFee(
		order.MerchantInfo.MerchantID,
		currency,
		amountFiat*order.PlatformFee,
	)
}

// recordOrderCompletedMetrics - вызывается при завершении заказа (COMPLETED)
func (uc *DefaultOrderUsecase) recordOrderCompletedMetrics(order *domain.Order, paymentSystem string) {
	if uc.Metrics == nil {
		return
	}

	orderType := string(order.Type)
	currency := order.AmountInfo.Currency
	amountFiat := order.AmountInfo.AmountFiat

	uc.Metrics.RecordOrderCompleted(
		order.MerchantInfo.MerchantID,
		orderType,
		currency,
		paymentSystem,
		amountFiat,
		order.RequisiteDetails.TraderID,
	)

	// Записываем награду трейдеру
	uc.Metrics.RecordTraderReward(
		order.RequisiteDetails.TraderID,
		currency,
		amountFiat*order.TraderReward,
	)

	// Записываем время обработки
	if !order.Metrics.CompletedAt.IsZero() && !order.CreatedAt.IsZero() {
		duration := order.Metrics.CompletedAt.Sub(order.CreatedAt).Seconds()
		uc.Metrics.RecordOrderProcessingDuration(
			order.MerchantInfo.MerchantID,
			string(domain.StatusCompleted),
			duration,
		)
	}
}

// recordOrderCanceledMetrics - вызывается при отмене заказа (CANCELED)
func (uc *DefaultOrderUsecase) recordOrderCanceledMetrics(order *domain.Order, paymentSystem string) {
	if uc.Metrics == nil {
		return
	}

	orderType := string(order.Type)
	currency := order.AmountInfo.Currency
	amountFiat := order.AmountInfo.AmountFiat

	uc.Metrics.RecordOrderCanceled(
		order.MerchantInfo.MerchantID,
		orderType,
		currency,
		paymentSystem,
		amountFiat,
	)

	// Записываем время обработки
	if !order.Metrics.CanceledAt.IsZero() && !order.CreatedAt.IsZero() {
		duration := order.Metrics.CanceledAt.Sub(order.CreatedAt).Seconds()
		uc.Metrics.RecordOrderProcessingDuration(
			order.MerchantInfo.MerchantID,
			string(domain.StatusCanceled),
			duration,
		)
	}
}

// recordOrderErrorMetrics - записывает ошибку
func (uc *DefaultOrderUsecase) recordOrderErrorMetrics(merchantID, errorType string) {
	if uc.Metrics == nil {
		return
	}

	uc.Metrics.RecordError(merchantID, errorType)
}