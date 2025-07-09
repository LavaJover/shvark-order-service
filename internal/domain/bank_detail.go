package domain

import "time"

type BankDetail struct {
	ID 						string
	TraderID 				string
	Country 				string
	Currency 				string
	InflowCurrency			string
	MinAmount 				float32
	MaxAmount 				float32
	BankName 				string
	PaymentSystem 			string
	Delay					time.Duration
	Enabled 				bool
	CardNumber 				string
	Phone 					string
	Owner 					string
	MaxOrdersSimultaneosly  int32
	MaxAmountDay			int32
	MaxAmountMonth  		int32
	MaxQuantityDay			int32
	MaxQuantityMonth		int32
	DeviceID				string
	CreatedAt 				time.Time
	UpdatedAt				time.Time
}

type BankDetailQuery struct {
	Amount 			float32
	Currency 		string
	PaymentSystem 	string
	Country 		string
}

type BankDetailFilters struct {
	DateFrom time.Time	`form:"date_from"`
	DateTo 	 time.Time	`form:"date_to"`
}

type BankDetailRepository interface {
	SaveBankDetail(bankDetail *BankDetail) error
	CreateBankDetail(bankDetail *BankDetail) (string, error)
	UpdateBankDetail(bankDetail *BankDetail) error
	DeleteBankDetail(bankDetailID string) error
	GetBankDetailByID(bankDetailID string) (*BankDetail, error)
	GetBankDetailsByTraderID(
		bankDetailID string,
		page, limit int,
		sortBy, sortOrder string,
	) ([]*BankDetail, int64, error)
}

type BankDetailUsecase interface {
	CreateBankDetail(bankDetail *BankDetail) (string, error)
	UpdateBankDetail(bankDetail *BankDetail) error
	DeleteBankDetail(bankDetailID string) error
	GetBankDetailByID(bankDetailID string) (*BankDetail, error)
	GetBankDetailsByTraderID(
		traderID string,
		page, limit int,
		sortBy, sortOrder string,
	) ([]*BankDetail, int64, error)
}