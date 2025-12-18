package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// OrderMetrics содержит все метрики для заказов
type OrderMetrics struct {
	// Счетчики создаваемых сделок
	OrdersCreatedTotal          prometheus.CounterVec
	OrdersCreatedAmountTotal    prometheus.CounterVec
	OrdersCreatedCount          prometheus.GaugeVec
	
	// Успешно завершенные сделки (COMPLETED)
	OrdersCompletedTotal        prometheus.CounterVec
	OrdersCompletedAmountTotal  prometheus.CounterVec
	OrdersCompletedCount        prometheus.GaugeVec
	
	// Отмененные сделки (CANCELED)
	OrdersCanceledTotal         prometheus.CounterVec
	OrdersCanceledAmountTotal   prometheus.CounterVec
	OrdersCanceledCount         prometheus.GaugeVec
	
	// Метрики по статусам
	OrderStatusGauge            prometheus.GaugeVec
	
	// Метрики по типам (PAYIN/PAYOUT)
	OrderTypeCreatedTotal       prometheus.CounterVec
	
	// Метрики по мерчантам
	MerchantOrdersCreatedTotal  prometheus.CounterVec
	MerchantOrdersCompletedTotal prometheus.CounterVec
	MerchantOrdersCanceledTotal prometheus.CounterVec
	MerchantAmountCreatedTotal  prometheus.CounterVec
	MerchantAmountCompletedTotal prometheus.CounterVec
	
	// Метрики по трейдерам
	TraderOrdersCompletedTotal  prometheus.CounterVec
	TraderAmountCompletedTotal  prometheus.CounterVec
	
	// Время обработки
	OrderProcessingDuration     prometheus.HistogramVec
	
	// Комиссии и награды
	PlatformFeeTotal            prometheus.CounterVec
	TraderRewardTotal           prometheus.CounterVec
	
	// Ошибки
	OrderErrorsTotal            prometheus.CounterVec
}

// NewOrderMetrics создает новый экземпляр метрик
func NewOrderMetrics() *OrderMetrics {
	return &OrderMetrics{
		// Созданные заказы (все)
		OrdersCreatedTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "orders_created_total",
				Help: "Общее количество созданных заказов",
			},
			[]string{"merchant_id", "order_type", "currency"},
		),
		OrdersCreatedAmountTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "orders_created_amount_total",
				Help: "Общая сумма созданных заказов в фиате",
			},
			[]string{"merchant_id", "currency"},
		),
		OrdersCreatedCount: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "orders_created_count",
				Help: "Текущее количество открытых заказов (PENDING/CREATED)",
			},
			[]string{"merchant_id"},
		),
		
		// Успешно завершенные заказы (COMPLETED)
		OrdersCompletedTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "orders_completed_total",
				Help: "Общее количество завершенных заказов (статус COMPLETED)",
			},
			[]string{"merchant_id", "order_type", "currency"},
		),
		OrdersCompletedAmountTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "orders_completed_amount_total",
				Help: "Общая сумма завершенных заказов в фиате",
			},
			[]string{"merchant_id", "currency"},
		),
		OrdersCompletedCount: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "orders_completed_count",
				Help: "Текущее количество завершенных заказов",
			},
			[]string{"merchant_id"},
		),
		
		// Отмененные заказы (CANCELED)
		OrdersCanceledTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "orders_canceled_total",
				Help: "Общее количество отмененных заказов (невыдача)",
			},
			[]string{"merchant_id", "order_type", "currency"},
		),
		OrdersCanceledAmountTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "orders_canceled_amount_total",
				Help: "Общая сумма отмененных заказов в фиате",
			},
			[]string{"merchant_id", "currency"},
		),
		OrdersCanceledCount: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "orders_canceled_count",
				Help: "Текущее количество отмененных заказов",
			},
			[]string{"merchant_id"},
		),
		
		// Статусы
		OrderStatusGauge: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "orders_by_status",
				Help: "Количество заказов по статусам",
			},
			[]string{"merchant_id", "status"},
		),
		
		// По типам
		OrderTypeCreatedTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "orders_by_type_total",
				Help: "Количество созданных заказов по типам (DEPOSIT/PAYOUT)",
			},
			[]string{"merchant_id", "order_type"},
		),
		
		// По мерчантам
		MerchantOrdersCreatedTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "merchant_orders_created_total",
				Help: "Общее количество заказов по мерчантам",
			},
			[]string{"merchant_id"},
		),
		MerchantOrdersCompletedTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "merchant_orders_completed_total",
				Help: "Общее количество завершенных заказов по мерчантам",
			},
			[]string{"merchant_id"},
		),
		MerchantOrdersCanceledTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "merchant_orders_canceled_total",
				Help: "Общее количество отмененных заказов по мерчантам",
			},
			[]string{"merchant_id"},
		),
		MerchantAmountCreatedTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "merchant_amount_created_total",
				Help: "Общая сумма созданных заказов по мерчантам",
			},
			[]string{"merchant_id", "currency"},
		),
		MerchantAmountCompletedTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "merchant_amount_completed_total",
				Help: "Общая сумма завершенных заказов по мерчантам",
			},
			[]string{"merchant_id", "currency"},
		),
		
		// По трейдерам
		TraderOrdersCompletedTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trader_orders_completed_total",
				Help: "Количество успешно завершенных заказов по трейдерам",
			},
			[]string{"trader_id"},
		),
		TraderAmountCompletedTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trader_amount_completed_total",
				Help: "Сумма успешно завершенных заказов по трейдерам",
			},
			[]string{"trader_id", "currency"},
		),
		
		// Время обработки
		OrderProcessingDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "order_processing_duration_seconds",
				Help:    "Время обработки заказа в секундах",
				Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1s, 2s, 4s, 8s...
			},
			[]string{"merchant_id", "status"},
		),
		
		// Комиссии
		PlatformFeeTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "platform_fee_total",
				Help: "Общая сумма комиссий платформы",
			},
			[]string{"merchant_id", "currency"},
		),
		TraderRewardTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trader_reward_total",
				Help: "Общая сумма награды трейдерам",
			},
			[]string{"trader_id", "currency"},
		),
		
		// Ошибки
		OrderErrorsTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_errors_total",
				Help: "Общее количество ошибок при создании/обработке заказов",
			},
			[]string{"merchant_id", "error_type"},
		),
	}
}

// RecordOrderCreated записывает созданный заказ
func (m *OrderMetrics) RecordOrderCreated(merchantID, orderType, currency string, amountFiat float64) {
	m.OrdersCreatedTotal.WithLabelValues(merchantID, orderType, currency).Inc()
	m.OrdersCreatedAmountTotal.WithLabelValues(merchantID, currency).Add(amountFiat)
	m.OrdersCreatedCount.WithLabelValues(merchantID).Inc()
	m.OrderTypeCreatedTotal.WithLabelValues(merchantID, orderType).Inc()
	m.MerchantOrdersCreatedTotal.WithLabelValues(merchantID).Inc()
	m.MerchantAmountCreatedTotal.WithLabelValues(merchantID, currency).Add(amountFiat)
}

// RecordOrderCompleted записывает завершенный заказ
func (m *OrderMetrics) RecordOrderCompleted(merchantID, orderType, currency string, amountFiat float64, traderID string) {
	m.OrdersCompletedTotal.WithLabelValues(merchantID, orderType, currency).Inc()
	m.OrdersCompletedAmountTotal.WithLabelValues(merchantID, currency).Add(amountFiat)
	m.OrdersCompletedCount.WithLabelValues(merchantID).Inc()
	m.MerchantOrdersCompletedTotal.WithLabelValues(merchantID).Inc()
	m.MerchantAmountCompletedTotal.WithLabelValues(merchantID, currency).Add(amountFiat)
	m.TraderOrdersCompletedTotal.WithLabelValues(traderID).Inc()
	m.TraderAmountCompletedTotal.WithLabelValues(traderID, currency).Add(amountFiat)
	m.OrdersCreatedCount.WithLabelValues(merchantID).Dec()
}

// RecordOrderCanceled записывает отмененный заказ
func (m *OrderMetrics) RecordOrderCanceled(merchantID, orderType, currency string, amountFiat float64) {
	m.OrdersCanceledTotal.WithLabelValues(merchantID, orderType, currency).Inc()
	m.OrdersCanceledAmountTotal.WithLabelValues(merchantID, currency).Add(amountFiat)
	m.OrdersCanceledCount.WithLabelValues(merchantID).Inc()
	m.MerchantOrdersCanceledTotal.WithLabelValues(merchantID).Inc()
	m.OrdersCreatedCount.WithLabelValues(merchantID).Dec()
}

// RecordOrderStatus обновляет статус заказа
func (m *OrderMetrics) RecordOrderStatus(merchantID, status string) {
	m.OrderStatusGauge.WithLabelValues(merchantID, status).Inc()
}

// RecordOrderProcessingDuration записывает время обработки
func (m *OrderMetrics) RecordOrderProcessingDuration(merchantID, finalStatus string, durationSeconds float64) {
	m.OrderProcessingDuration.WithLabelValues(merchantID, finalStatus).Observe(durationSeconds)
}

// RecordPlatformFee записывает комиссию платформы
func (m *OrderMetrics) RecordPlatformFee(merchantID, currency string, feeAmount float64) {
	m.PlatformFeeTotal.WithLabelValues(merchantID, currency).Add(feeAmount)
}

// RecordTraderReward записывает награду трейдеру
func (m *OrderMetrics) RecordTraderReward(traderID, currency string, rewardAmount float64) {
	m.TraderRewardTotal.WithLabelValues(traderID, currency).Add(rewardAmount)
}

// RecordError записывает ошибку
func (m *OrderMetrics) RecordError(merchantID, errorType string) {
	m.OrderErrorsTotal.WithLabelValues(merchantID, errorType).Inc()
}