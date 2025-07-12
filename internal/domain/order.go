package domain

import (
	"context"
	"time"
)

type OrderStatus string

const (
	StatusCreated 		  OrderStatus = "CREATED"
	StatusPreorder		  OrderStatus = "PREORDER_CREATED"
	StatusFailed 		  OrderStatus = "FAILED"
	StatusCanceled 		  OrderStatus = "CANCELED"
	StatusSucceed 		  OrderStatus = "SUCCEED"
	StatusDisputeCreated  OrderStatus = "DISPUTE_CREATED"
)

type Order struct {
	ID 			  		string
	MerchantID 	  		string
	AmountFiat 	  		float64
	AmountCrypto  		float64
	Currency 	  		string
	Country 	  		string
	ClientID   			string
	Status 		  		OrderStatus
	PaymentSystem 		string
	BankDetailsID 		string
	BankDetail    		*BankDetail
	ExpiresAt	  		time.Time
	CreatedAt 	  		time.Time
	UpdatedAt 	  		time.Time
	MerchantOrderID 	string
	Shuffle 			int32
	CallbackURL 		string
	TraderRewardPercent float64
	PlatformFee 		float64
	Recalculated 		bool
	CryptoRubRate		float64
	BankCode 			string
	NspkCode 			string
}

type OrderFilters struct {
	Statuses 		[]string  `form:"status"`
	MinAmountFiat 	float64   `form:"min_amount"`
	MaxAmountFiat 	float64	  `form:"max_amount"`
	DateFrom 		time.Time `form:"date_from"`
	DateTo 			time.Time `form:"date_to"`
	Currency 		string 	  `form:"currency"`
}

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

type OrderRepository interface {
	CreateOrder(order *Order) (string, error)
	UpdateOrderStatus(orderID string, newStatus OrderStatus) error
	GetOrderByID(orderID string) (*Order, error)
	GetOrdersByTraderID(
		orderID string, page, 
		limit int64, sortBy, 
		sortOrder string, 
		filters OrderFilters,
		) ([]*Order, int64, error)
	GetOrdersByBankDetailID(bankDetailID string) ([]*Order, error)
	FindExpiredOrders() ([]*Order, error)
	GetCreatedOrdersByClientID(clientID string) ([]*Order, error)
}