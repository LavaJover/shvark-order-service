package orderdto

import "time"

type CreatePayOutOrderInput struct {
	MerchantParams
	AdvancedParams
	Type string
	ExpiresAt time.Time
}

type PaymentDetails struct {
	PaymentSystem 	string
	CardNumber		string
	Phone			string
	Currency		string
	UsdRate			float64
	BankInfo
}