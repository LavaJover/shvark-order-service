package domain

import "time"

type UncreatedOrder struct {
	ID              string
	MerchantID      string
	AmountFiat      float64
	AmountCrypto    float64
	Currency        string
	ClientID        string
	CreatedAt       time.Time
	MerchantOrderID string
	PaymentSystem   string
	BankCode        string
	ErrorMessage    string
}

type UncreatedOrdersFilter struct {
	MerchantID       *string
	MinAmountFiat    *float64
	MaxAmountFiat    *float64
	Currency         *string
	ClientID         *string
	TimeOpeningStart *time.Time
	TimeOpeningEnd   *time.Time
	PaymentSystem    *string
	BankCode         *string
}

type UncreatedOrdersStats struct {
	MerchantID    *string
	Currency      *string
	PaymentSystem *string
	DateGroup     *string
	AmountRange   *string
	BankCode      *string

	TotalCount      int64
	TotalAmountFiat float64
	AvgAmountFiat   float64
	MinAmountFiat   float64
	MaxAmountFiat   float64
}

type UncreatedOrdersResponseStats struct {
	TotalCount        int64
	TotalAmountFiat   float64
	TotalAmountCrypto float64
	AvgAmountFiat     float64
	AvgAmountCrypto   float64
}

type UncreatedOrderRepository interface {
	CreateLog(log *UncreatedOrder) error

	GetLogsWithFilters(filter *UncreatedOrdersFilter, page, limit int32, sortBy, sortOrder string) ([]*UncreatedOrder, int64, error)
	GetStatsWithFIlters(filter *UncreatedOrdersFilter, groupByCriteria []string) ([]*UncreatedOrdersStats, error)
}
