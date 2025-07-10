package models

import (
	"time"

	"gorm.io/gorm"
)

type BankDetailModel struct {
	ID 						string	`gorm:"primaryKey;type:uuid"`
	TraderID 				string	`gorm:"type:uuid;not null"`
	Country 				string	
	Currency 				string
	InflowCurrency 			string
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
	MaxQuantityDay			int32
	MaxQuantityMonth		int32
	DeviceID				string
	CreatedAt				time.Time
	UpdatedAt 				time.Time
	DeletedAt 				gorm.DeletedAt `gorm:"index"`
}