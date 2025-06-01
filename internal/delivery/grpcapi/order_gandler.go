package grpcapi

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/client"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
)

type OrderHandler struct {
	uc *domain.OrderUsecase
	client client.BankingClient
}

func NewOrderHandler(uc *domain.OrderUsecase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, r *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	// find bank details
	h.client
}