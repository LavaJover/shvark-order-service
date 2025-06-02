package postgres

import "time"

type OrderModel struct {
	ID 			  string  `gorm:"primaryKey;type:uuid"`
	MerchantID 	  string  `gorm:"type:uuid"`
	Amount 		  float32	
	Currency 	  string		
	Country 	  string
	ClientEmail   string
	MetadataJSON  string
	Status 		  string
	PaymentSystem string
	BankDetailsID string  `gorm:"type:uuid"`	
}

type BankDetailModel struct {
	ID 						string	`gorm:"primaryKey;type:uuid"`
	TraderID 				string	`gorm:"type:uuid;not null"`
	Country 				string	
	Currency 				string	
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
	MaxAmountMonth			int32
}