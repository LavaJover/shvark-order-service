package domain

import "time"

type OrderStatus string

const (
	StatusCreated 		  OrderStatus = "CREATED"
	StatusCanceled 		  OrderStatus = "CANCELED"
	StatusSucceed 		  OrderStatus = "SUCCEED"
	StatusDisputeCreated  OrderStatus = "DISPUTE_CREATED"
	StatusDisputeResolved OrderStatus = "DISPUTE_RESOLVED"
)

type Order struct {
	ID 			  		string
	MerchantID 	  		string
	AmountFiat 	  		float64
	AmountCrypto  		float64
	Currency 	  		string
	Country 	  		string
	ClientID   			string
	Status 		  		OrderStatus
	PaymentSystem 		string
	BankDetailsID 		string
	BankDetail    		*BankDetail
	ExpiresAt	  		time.Time
	CreatedAt 	  		time.Time
	UpdatedAt 	  		time.Time
	MerchantOrderID 	string
	Shuffle 			int32
	CallbackURL 		string
	TraderRewardPercent float64
	Recalculated 		bool
}

type OrderFilters struct {
	Statuses 		[]string  `form:"status"`
	MinAmountFiat 	float64   `form:"min_amount"`
	MaxAmountFiat 	float64	  `form:"max_amount"`
	DateFrom 		time.Time `form:"date_from"`
	DateTo 			time.Time `form:"date_to"`
	Currency 		string 	  `form:"currency"`
}