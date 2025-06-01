package domain

type Order struct {
	ID 			  string
	MerchantID 	  string
	Amount 		  float32
	Currency 	  string
	Country 	  string
	ClientEmail   string
	MetadataJSON  string
	Status 		  string
	PaymentSystem string
	BankDetailsID string
}