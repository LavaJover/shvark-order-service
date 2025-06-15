package postgres

import (
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
)

type OrderModel struct {
	ID 			  	string  			`gorm:"primaryKey;type:uuid"`
	MerchantID 	  	string  			
	AmountFiat 	  	float64
	AmountCrypto  	float64	
	Currency 	  	string		
	Country 	  	string
	ClientID   	  	string
	Status 		  	domain.OrderStatus	`gorm:"index:idx_status_expires"`
	PaymentSystem 	string
	BankDetailsID 	string  			`gorm:"type:uuid"`	
	BankDetail 	  	BankDetailModel   	`gorm:"foreignKey:BankDetailsID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ExpiresAt  	  	time.Time			`gorm:"index:idx_status_expires"`
	CreatedAt 	  	time.Time
	UpdatedAt 	  	time.Time
	MerchantOrderID string
	Shuffle 		int32
	CallbackURL 	string
}