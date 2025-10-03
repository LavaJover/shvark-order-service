package models

import (
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
)

type OrderModel struct {
	ID 			  		string  			`gorm:"primaryKey;type:uuid"`
	MerchantID 	  		string  			
	AmountFiat 	  		float64				`gorm:"index:idx_amount"`
	AmountCrypto  		float64	
	Currency 	  		string		
	Country 	  		string
	ClientID   	  		string
	Status 		  		domain.OrderStatus	`gorm:"index:idx_status_expires"`
	BankDetailsID 		string  			`gorm:"type:uuid"`	
	BankDetail 	  		BankDetailModel   	`gorm:"foreignKey:BankDetailsID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ExpiresAt  	  		time.Time			`gorm:"index:idx_status_expires"`
	CreatedAt 	  		time.Time			`gorm:"index:idx_created_at"`
	UpdatedAt 	  		time.Time
	MerchantOrderID 	string
	Shuffle 			int32
	CallbackURL 		string
	TraderRewardPercent float64
	PlatformFee 		float64
	Recalculated   		bool
	CryptoRubRate		float64
	Type 				string
}

type PaymentProcessingLog struct {
	ID           string    `gorm:"primaryKey;type:uuid"`
	OrderID      string    `gorm:"type:uuid;not null;index"`
	PaymentHash  string    `gorm:"not null;index"` // Хэш уведомления для идемпотентности
	Amount       float64   `gorm:"not null"`
	PaymentSystem string   `gorm:"not null"`
	ProcessedAt  time.Time `gorm:"not null"`
	Success      bool      `gorm:"not null"`
	Error        string    
	Metadata     string    `gorm:"type:jsonb"` // Дополнительные данные
}

type AutomaticPaymentResult struct {
	Action  string                  `json:"action"`
	Message string                  `json:"message"`
	Results []OrderProcessingResult `json:"results,omitempty"`
}

type OrderProcessingResult struct {
	OrderID string `json:"order_id"`
	Action  string `json:"action"` // approved, failed, already_processed
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}