package kafka

type DisputeEvent struct {
	DisputeID string 	`json:"dispute_id"`
	OrderID 	string	`json:"order_id"`
	TraderID 	string  `json:"trader_id"`
	TraderName  string  `json:"trader_name"`
	ProofUrl 	string	`json:"proof_url"`
	Reason 		string	`json:"reason"`
	Status 		string	`json:"status"`
}