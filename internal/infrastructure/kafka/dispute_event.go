package kafka

type DisputeEvent struct {
	DisputeID 			string 	`json:"dispute_id"`
	OrderID 			string	`json:"order_id"`
	TraderID 			string  `json:"trader_id"`
	ProofUrl 			string	`json:"proof_url"`
	Reason 				string	`json:"reason"`
	Status 				string	`json:"status"`
	OrderAmountFiat 	float64 `json:"order_amount"`
	DisputeAmountFiat 	float64 `json:"dispute_amount_fiat"`
	BankName 			string  `json:"bank_name"`
	Phone 				string  `json:"phone"`
	CardNumber 			string  `json:"card_number"`
	Owner 				string  `json:"owner"`
}