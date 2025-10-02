package notifier

import "time"

type CallbackPayload struct {
	OrderID 		string 		`json:"order_id"`
	MerchantOrderID string 		`json:"merchant_order_id"`
	Status 			string 		`json:"status"`
	AmountFiat 		float64		`json:"amount64"`
	AmountCrypto 	float64		`json:"amount_crypto"`
	Currency 		string 		`json:"currency"`
	ConfirmedAt 	time.Time	`json:"confirmed_at"`
	ClientID		string 		`json:"client_id"`
}