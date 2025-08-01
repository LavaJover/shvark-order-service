package bankdetaildto

import "time"

type CreateBankDetailInput struct{
	SearchParams
	DeviceInfo
	TraderInfo
	PaymentDetails
	Country 		string
	Currency 		string
	InflowCurrency 	string
}

type SearchParams struct {
	MaxOrdersSimultaneosly  int32
	MaxAmountDay			float64
	MaxAmountMonth  		float64
	MaxQuantityDay			int32
	MaxQuantityMonth		int32
	MinOrderAmount 			float32
	MaxOrderAmount 			float32
	Delay 					time.Duration
	Enabled 				bool
}

type DeviceInfo struct {
	DeviceID string
}

type TraderInfo struct {
	TraderID string
}

type BankInfo struct {
	BankCode string
	BankName string
	NspkCode string
}

type PaymentDetails struct {
	Phone 			string
	CardNumber 		string
	Owner 			string
	PaymentSystem 	string
	BankInfo
}

type UpdateBankDetailInput struct {
	ID              string
	SearchParams
	DeviceInfo
	TraderInfo
	PaymentDetails
	Country 		string
	Currency 		string
	InflowCurrency 	string
}

type FindSuitableBankDetailsInput struct {
	AmountFiat float64
	Currency string
	PaymentSystem string
	BankCode string
	NspkCode string
}