package request

type ReleaseRequest struct {
	TraderID  	  string  `json:"traderId"`
	OrderID  	  string  `json:"orderId"`
	MerchantID	  string  `json:"merchantId"`
	PlatformFee   float64  `json:"platformFee"`
	RewardPercent float64 `json:"rewardPercent"`
	CommissionUsers []CommissionUser `json:"commissionUsers,omitempty"`
}

type CommissionUser struct {
	UserID string `json:"userId"`
	Commission float64 `json:"commission"`
}