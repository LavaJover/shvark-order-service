package domain

type OrderUsecase interface {
	CreateOrder(order *Order) (*Order, error)
	ApproveOrder(orderID string) error
	CancelOrder(orderID string) error
	GetOrderByID(orderID string) (*Order, error)
	GetOrdersByTraderID(orderID string) ([]*Order, error)
}