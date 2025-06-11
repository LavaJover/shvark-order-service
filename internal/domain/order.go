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
	ID 			  string
	MerchantID 	  string
	AmountFiat 	  float64
	AmountCrypto  float64
	Currency 	  string
	Country 	  string
	ClientEmail   string
	MetadataJSON  string
	Status 		  OrderStatus
	PaymentSystem string
	BankDetailsID string
	BankDetail    *BankDetail
	ExpiresAt	  time.Time
}