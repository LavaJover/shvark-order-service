package request

type ReleaseRequest struct {
	TraderID  	  string  `json:"traderId"`
	OrderID  	  string  `json:"orderId"`
	MerchantID	  string  `json:"merchantId"`
	PlatformFee   float64  `json:"platformFee"`
	RewardPercent float64 `json:"rewardPercent"`
}