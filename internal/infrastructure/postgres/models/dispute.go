package models

import (
	"time"
)

type DisputeModel struct {
	ID 			 		string `gorm:"primaryKey"`
	OrderID 	 		string
	OrderStatusOriginal string
	OrderStatusDisputed string
	DisputeAmountFiat 	float64
	DisputeAmountCrypto float64
	DisputeCryptoRate 	float64
	ProofUrl 	 		string
	Reason 		 		string
	Status 		 		string
	Order		 		OrderModel `gorm:"foreignKey:OrderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CreatedAt	 		time.Time
	UpdatedAt 	 		time.Time
	Ttl					time.Duration
	AutoAcceptAt 		time.Time   
}