package domain

import "context"

type OrderUsecase interface {
	CreateOrder(order *Order) (*Order, error)
	GetOrderByID(orderID string) (*Order, error)
	GetOrdersByTraderID(
		orderID string, page, 
		limit int64, sortBy, 
		sortOrder string, 
		filters OrderFilters,
		) ([]*Order, int64, error)
	FindExpiredOrders() ([]*Order, error)
	CancelExpiredOrders(context.Context) error
	OpenOrderDispute(orderID string) error
	ResolveOrderDispute(orderID string) error
	ApproveOrder(orderID string) error
	CancelOrder(orderID string) error
}