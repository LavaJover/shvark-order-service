package postgres

import (
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
)

type OrderModel struct {
	ID 			  string  			`gorm:"primaryKey;type:uuid"`
	MerchantID 	  string  			`gorm:"type:uuid"`
	AmountFiat 	  float64
	AmountCrypto  float64	
	Currency 	  string		
	Country 	  string
	ClientEmail   string
	MetadataJSON  string
	Status 		  domain.OrderStatus
	PaymentSystem string
	BankDetailsID string  			`gorm:"type:uuid"`	
	BankDetail 	  BankDetailModel   `gorm:"foreignKey:BankDetailsID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ExpiresAt  	  time.Time
}