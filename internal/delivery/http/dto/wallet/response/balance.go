package response

type BalanceResponse struct {
	TraderID string `json:"traderId"`
	Currency string `json:"currency"`
	Balance float64 `json:"balance"`
	Frozen	float64 `json:"frozen"`
	Address string `json:"address"`
}

type AllBalancesResponse struct {
	Success  bool                     `json:"success"`
	Count    int                      `json:"count"`
	Balances map[string]BalanceResponse `json:"balances"`
}

type BatchBalancesResponse struct {
	Success  bool              `json:"success"`
	Balances []BalanceResponse `json:"balances"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}