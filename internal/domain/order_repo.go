package domain

type OrderRepository interface {
	CreateOrder(order *Order) (string, error)
	UpdateOrderStatus(orderID string, newStatus OrderStatus) error
	GetOrderByID(orderID string) (*Order, error)
	GetOrdersByTraderID(orderID string, page, limit int64, sortBy, sortOrder string) ([]*Order, int64, error)
	GetOrdersByBankDetailID(bankDetailID string) ([]*Order, error)
	FindExpiredOrders() ([]*Order, error)
}