package domain

type OrderRepository interface {
	CreateOrder(order *Order) (string, error)
}