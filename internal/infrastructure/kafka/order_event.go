package publisher

type OrderEvent struct {
	OrderID 	string	`json:"order_id"`
	TraderID 	string	`json:"trader_id"`
	Status 		string	`json:"status"`
	AmountFiat 	float64	`json:"amount_fiat"`
	Currency 	string	`json:"currency"`
	BankName    string  `json:"bank_name"`
	Phone  		string  `json:"phone"`
	CardNumber  string  `json:"card_number"`
	Owner       string  `json:"owner"`
}