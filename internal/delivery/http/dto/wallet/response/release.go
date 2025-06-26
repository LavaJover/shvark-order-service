package response

type ReleaseResponse struct {
	Released 	   float64 `json:"released"`
	Reward   	   float64 `json:"reward"`
	Platform 	   float64 `json:"platform"`
	MerchantAmount float64 `json:"merchant_amount"`
}