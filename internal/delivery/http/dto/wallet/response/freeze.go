package response

type FreezeResponse struct {
	Frozen float64 `json:"frozen"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}