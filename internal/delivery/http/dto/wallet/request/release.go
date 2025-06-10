package request

type ReleaseRequest struct {
	TraderID  	  string  `json:"trader_id"`
	OrderID  	  string  `json:"order_id"`
	RewardPercent float64 `json:"rewardPercent"`
}