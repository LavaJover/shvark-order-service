package domain

type OrderUsecase interface {
	CreateOrder(order *Order) (string, error)
}