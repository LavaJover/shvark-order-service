package request

type ReleaseRequest struct {
	TraderID  	  string  `json:"traderId"`
	OrderID  	  string  `json:"orderId"`
	RewardPercent float64 `json:"rewardPercent"`
}