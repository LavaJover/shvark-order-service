package kafka

type OrderEvent struct {
	OrderID 	string	`json:"order_id"`
	TraderID 	string	`json:"trader_id"`
	Status 		string	`json:"status"`
	AmountFiat 	float64	`json:"amount_fiat"`
	Currency 	string	`json:"currency"`
}