package domain

type OrderRepository interface {
	CreateOrder(order *Order) (string, error)
	UpdateOrderStatus(orderID string, newStatus string) error
	GetOrderByID(orderID string) (*Order, error)
	GetOrdersByTraderID(traderID string) ([]*Order, error)
}