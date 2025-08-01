package domain

import "time"

// type BankDetail struct {
// 	ID 						string +
// 	TraderID 				string +
// 	Country 				string +
// 	Currency 				string +
// 	InflowCurrency			string +
// 	MinAmount 				float32 +
// 	MaxAmount 				float32 +
// 	BankName 				string +
// 	BankCode 				string +
// 	NspkCode 				string +
// 	PaymentSystem 			string +
// 	Delay					time.Duration +
// 	Enabled 				bool +
// 	CardNumber 				string +
// 	Phone 					string +
// 	Owner 					string +
// 	MaxOrdersSimultaneosly  int32 +
// 	MaxAmountDay			int32+
// 	MaxAmountMonth  		int32+
// 	MaxQuantityDay			int32+
// 	MaxQuantityMonth		int32+
// 	DeviceID				string +
// 	CreatedAt 				time.Time +
// 	UpdatedAt				time.Time +
// }

type BankDetail struct {
	ID 				string
	SearchParams
	DeviceInfo
	TraderInfo
	PaymentDetails
	Country 		string
	Currency 		string
	InflowCurrency 	string
	CreatedAt 		time.Time
	UpdatedAt 		time.Time
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

type BankDetailQuery struct {
	Amount 			float32
	Currency 		string
	PaymentSystem 	string
	Country 		string
}

type BankDetailStat struct {
	BankDetailID   	   string  
	CurrentCountToday  int     
	CurrentCountMonth  int     
	CurrentAmountToday float64
	CurrentAmountMonth float64
}

type BankDetailFilters struct {
	DateFrom time.Time	`form:"date_from"`
	DateTo 	 time.Time	`form:"date_to"`
}

type SuitablleBankDetailsQuery struct {
	AmountFiat float64
	BankCode string
	NspkCode string
	PaymentSystem string
	Currency string
}

type BankDetailRepository interface {
	SaveBankDetail(bankDetail *BankDetail) error
	CreateBankDetail(bankDetail *BankDetail) error
	UpdateBankDetail(bankDetail *BankDetail) error
	DeleteBankDetail(bankDetailID string) error
	GetBankDetailByID(bankDetailID string) (*BankDetail, error)
	GetBankDetailsByTraderID(
		bankDetailID string,
		page, limit int,
		sortBy, sortOrder string,
	) ([]*BankDetail, int64, error)
	FindSuitableBankDetails(query *SuitablleBankDetailsQuery) ([]*BankDetail, error)
	GetBankDetailsStatsByTraderID(traderID string) ([]*BankDetailStat, error)
}