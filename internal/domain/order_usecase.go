package domain

type OrderUsecase interface {
	CreateOrder(order *Order) (*Order, error)
	GetOrderByID(orderID string) (*Order, error)
	GetOrdersByTraderID(orderID string) ([]*Order, error)
	FindExpiredOrders() ([]*Order, error)
	CancelExpiredOrders() error
	UpdateOrderStatus(orderID string, newStatus OrderStatus) error
}