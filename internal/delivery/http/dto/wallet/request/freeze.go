package request

type FreezeRequest struct {
	TraderID string  `json:"traderId"`
	OrderID  string  `json:"orderId"`
	Amount 	 float64 `json:"amount"`
}