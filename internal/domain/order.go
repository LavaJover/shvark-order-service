package domain

type OrderStatus string

const (
	StatusCreated 		  OrderStatus = "CREATED"
	StatusCanceled 		  OrderStatus = "CANCELED"
	StatusDisputeCreated  OrderStatus = "DISPUTE_CREATED"
	StatusDisputeResolved OrderStatus = "DISPUTE_RESOLVED"
)


type Order struct {
	ID 			  string
	MerchantID 	  string
	Amount 		  float32
	Currency 	  string
	Country 	  string
	ClientEmail   string
	MetadataJSON  string
	Status 		  OrderStatus
	PaymentSystem string
	BankDetailsID string
	BankDetail    *BankDetail
}