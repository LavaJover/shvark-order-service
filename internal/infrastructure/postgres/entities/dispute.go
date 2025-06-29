package entities

import "time"

type DisputeModel struct {
	ID 			 string `gorm:"primaryKey"`
	OrderID 	 string	
	ProofUrl 	 string
	Reason 		 string
	Status 		 string
	Order		 OrderModel `gorm:"foreignKey:OrderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CreatedAt	 time.Time
	UpdatedAt 	 time.Time
	AutoAcceptAt time.Time   
}