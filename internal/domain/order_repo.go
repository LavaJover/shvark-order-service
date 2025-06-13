package domain

type OrderRepository interface {
	CreateOrder(order *Order) (string, error)
	UpdateOrderStatus(orderID string, newStatus OrderStatus) error
	GetOrderByID(orderID string) (*Order, error)
	GetOrdersByTraderID(traderID string) ([]*Order, error)
	GetOrdersByBankDetailID(bankDetailID string) ([]*Order, error)
	FindExpiredOrders() ([]*Order, error)
}