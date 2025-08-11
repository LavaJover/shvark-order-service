package domain

import (
	"time"
)

type OrderStatus string

const (
	StatusPending 		  OrderStatus = "PENDING"
	StatusCreated		  OrderStatus = "CREATED"
	StatusFailed 		  OrderStatus = "FAILED"
	StatusCanceled 		  OrderStatus = "CANCELED"
	StatusCompleted 	  OrderStatus = "COMPLETED"
	StatusDisputeCreated  OrderStatus = "DISPUTE"
)

// type Order struct {
// 	ID 			  		string +
// 	MerchantID 	  		string +
// 	AmountFiat 	  		float64 +
// 	AmountCrypto  		float64 +
// 	Currency 	  		string +
// 	Country 	  		string - 
// 	ClientID   			string +
// 	Status 		  		OrderStatus +
// 	PaymentSystem 		string +
// 	BankDetailsID 		string +
// 	BankDetail    		*BankDetail
// 	ExpiresAt	  		time.Time +
// 	CreatedAt 	  		time.Time +
// 	UpdatedAt 	  		time.Time +
// 	MerchantOrderID 	string +
// 	Shuffle 			int32 +
// 	CallbackURL 		string +
// 	TraderRewardPercent float64 +
// 	PlatformFee 		float64 +
// 	Recalculated 		bool +
// 	CryptoRubRate		float64 +
// 	BankCode 			string +
// 	NspkCode 			string +
// 	Type 				string +
// }

type Order struct {
	ID 				string
	Status 			OrderStatus
	MerchantInfo	MerchantInfo
	AmountInfo 		AmountInfo
	BankDetailID 	string
	Type 			string
	Recalculated 	bool
	Shuffle 		int32
	TraderReward 	float64
	PlatformFee		float64
	CallbackUrl 	string

	ExpiresAt 		time.Time
	CreatedAt 		time.Time
	UpdatedAt 		time.Time
}

type MerchantInfo struct {
	MerchantID 		string
	MerchantOrderID string
	ClientID 		string
}

type AmountInfo struct {
	AmountFiat 		float64
	AmountCrypto 	float64
	CryptoRate 		float64
	Currency 		string
}

type OrderFilters struct {
	Statuses 		[]string  `form:"status"`
	MinAmountFiat 	float64   `form:"min_amount"`
	MaxAmountFiat 	float64	  `form:"max_amount"`
	DateFrom 		time.Time `form:"date_from"`
	DateTo 			time.Time `form:"date_to"`
	Currency 		string 	  `form:"currency"`
	OrderID			string    `form:"order_id"`
	MerchantOrderID string    `form:"merchant_order_id"`
}

type OrderStatistics struct {
	TotalOrders 			int64
	SucceedOrders 			int64
	CanceledOrders 			int64
	ProcessedAmountFiat 	float64
	ProcessedAmountCrypto 	float64
	CanceledAmountFiat 		float64
	CanceledAmountCrypto 	float64
	IncomeCrypto 			float64
}

type Filter struct {
	DealID  *string
	Type             *string
	Status           *string
	TimeOpeningStart *time.Time
	TimeOpeningEnd   *time.Time
	AmountMin        *float64
	AmountMax        *float64
	MerchantID       string
}

type AllOrdersFilters struct {
	TraderID 			string
	MerchantID 			string
	OrderID				string
	MerchantOrderID 	string
	Status 				string
	BankCode 			string
	TimeOpeningStart 	time.Time
	TimeOpeningEnd 		time.Time
	AmountFiatMin 		float64
	AmountFiatMax 		float64
	Type 				string
	DeviceID 			string
	PaymentSystem		string
}

type OrderRepository interface {
	CreateOrder(order *Order) error
	UpdateOrderStatus(orderID string, newStatus OrderStatus) error
	GetOrderByID(orderID string) (*Order, error)
	GetOrderByMerchantOrderID(merchantOrderID string) (*Order, error)
	GetOrdersByTraderID(
		orderID string, page, 
		limit int64, sortBy, 
		sortOrder string, 
		filters OrderFilters,
		) ([]*Order, int64, error)
	GetOrdersByBankDetailID(bankDetailID string) ([]*Order, error)
	FindExpiredOrders() ([]*Order, error)
	GetCreatedOrdersByClientID(clientID string) ([]*Order, error)
	GetOrderStatistics(traderID string, dateFrom, dateTo time.Time) (*OrderStatistics, error)

	GetOrders(filter Filter, sortField string, page, size int) ([]*Order, int64, error)

	GetAllOrders(filter *AllOrdersFilters, sort string, page, limit int32) ([]*Order, int64, error)
}