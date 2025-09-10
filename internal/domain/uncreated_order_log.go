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

type UncreatedOrdersOverallStats struct {
	TotalCount        int64
	TotalAmountFiat   float64
	TotalAmountCrypto float64
	AvgAmountFiat     float64
	AvgAmountCrypto   float64
	UniqueMerchants   int64
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
	//GetOverallStats()
	//GetStatsByMerchantID(merchantID string)
	//GetStatsByMerchantID(filter *UncreatedOrdersFilter)
}
