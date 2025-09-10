package uncreatedorderdto

import "github.com/LavaJover/shvark-order-service/internal/domain"

type GetUncreatedOrdersOutput struct {
	UncreatedOrders []*domain.UncreatedOrder
	Pagination      Pagination
}

type Pagination struct {
	CurrentPage  int32
	TotalPages   int32
	TotalItems   int32
	ItemsPerPage int32
}
