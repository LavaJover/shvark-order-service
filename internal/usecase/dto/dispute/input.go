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