package disputedto

import "time"

type CreateDisputeInput struct {
	OrderID 		  	string
	DisputeAmountFiat 	float64
	DisputeAmountCrypto float64
	DisputeCryptoRate 	float64
	ProofUrl 			string
	Reason 				string
	Ttl					time.Duration
}

type GetOrderDisputesInput struct {
	Page 		int64
	Limit 		int64
	Status 		*string
	TraderID 	*string
	DisputeID 	*string
	MerchantID 	*string
	OrderID 	*string
}