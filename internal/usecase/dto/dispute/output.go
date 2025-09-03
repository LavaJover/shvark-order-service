package disputedto

import "github.com/LavaJover/shvark-order-service/internal/domain"

type GetOrderDisputesOutput struct {
	Disputes []*domain.Dispute
	Pagination Pagination
}

type Pagination struct {
	CurrentPage int32
	TotalPages	int32
	TotalItems	int32
	ItemsPerPage int32
}