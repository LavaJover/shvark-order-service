package models

import (
	"time"
)

type UncreatedOrderModel struct {
	ID              string  `gorm:"primaryKey;type:uuid"`
	MerchantID      string  `gorm:"index:idx_merchant_id_uncreated"`
	AmountFiat      float64 `gorm:"index:idx_amount_uncreated"`
	AmountCrypto    float64
	Currency        string
	ClientID        string
	CreatedAt       time.Time `gorm:"index:idx_created_at_uncreated"`
	MerchantOrderID string
	PaymentSystem   string
	BankCode        string
	ErrorMessage    string
}
