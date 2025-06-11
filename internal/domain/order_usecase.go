package domain

import "context"

type OrderUsecase interface {
	CreateOrder(order *Order) (*Order, error)
	GetOrderByID(orderID string) (*Order, error)
	GetOrdersByTraderID(orderID string) ([]*Order, error)
	FindExpiredOrders() ([]*Order, error)
	CancelExpiredOrders(context.Context) error
	OpenOrderDispute(orderID string) error
	ResolveOrderDispute(orderID string) error
	ApproveOrder(orderID string) error
	CancelOrder(orderID string) error
}