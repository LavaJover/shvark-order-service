package orderdto

import "time"

type CreatePayOutOrderInput struct {
	MerchantParams
	AdvancedParams
	PaymentDetails
	Type string
	ExpiresAt time.Time
}

type PaymentDetails struct {
	PaymentSystem 	string
	CardNumber		string
	Phone			string
	Owner			string
	Currency		string
	UsdRate			float64
	AmoutFiat		float64
	BankInfo
}