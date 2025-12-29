package domain

import (
	"context"
	"time"
)

type OrderType string

const (
	TypePayIn OrderType = "DEPOSIT"
	TypePayOut OrderType = "PAYOUT"
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

type Order struct {
	ID 				string
	Status 			OrderStatus
	MerchantInfo	MerchantInfo
	AmountInfo 		AmountInfo
	BankDetailID 	*string
	Type 			OrderType
	Recalculated 	bool
	Shuffle 		int32
	TraderReward 	float64
	PlatformFee		float64
	CallbackUrl 	string

	// Параметры реквизита
	RequisiteDetails RequisiteDetails
	// Метрики
	Metrics Metrics

	ExpiresAt 		time.Time
	CreatedAt 		time.Time
	UpdatedAt 		time.Time
}

type RequisiteDetails struct {
	TraderID 			string
	CardNumber 			string 
	Phone 				string
	Owner 				string
	PaymentSystem 		string
	BankName 			string
	BankCode 			string
	NspkCode 			string
	DeviceID			string
}

type Metrics struct {
	AutomaticCompleted  bool
	ManuallyCompleted   bool

	CompletedAt		time.Time
	CanceledAt		time.Time
	AcceptedAt 		time.Time
}

type MerchantInfo struct {
	MerchantID 		string
	MerchantOrderID string
	ClientID 		string
	StoreID			string
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

	ProcessOrderCriticalOperation(
		orderID string, 
		newStatus OrderStatus, 
		operation string, // добавляем параметр операции
		walletFunc func() error,
	) error

	CheckDuplicatePayment(ctx context.Context, orderID string, paymentHash string) (bool, error)
	FindPendingOrdersByDeviceID(deviceID string) ([]*Order, error)

	SaveAutomaticLog(ctx context.Context, log *AutomaticLog) error
    GetAutomaticLogs(ctx context.Context, filter *AutomaticLogFilter) ([]*AutomaticLog, error)

	GetAutomaticLogsCount(ctx context.Context, filter *AutomaticLogFilter) (int64, error)
	GetAutomaticStats(ctx context.Context, traderID string, days int) (*AutomaticStats, error)

	// Методы для транзакций
	BeginTx() (OrderRepository, error)
	Commit() error
	Rollback() error
	
	// Методы для работы в транзакции
	CreateOrderInTx(order *Order) error
	GetCreatedOrdersByClientIDInTx(clientID string) ([]*Order, error)
}

type PaymentProcessingLog struct {
	ID           string    `gorm:"primaryKey;type:uuid"`
	OrderID      string    `gorm:"type:uuid;not null;index"`
	PaymentHash  string    `gorm:"not null;index"` // Хэш уведомления для идемпотентности
	Amount       float64   `gorm:"not null"`
	PaymentSystem string   `gorm:"not null"`
	ProcessedAt  time.Time `gorm:"not null"`
	Success      bool      `gorm:"not null"`
	Error        string    
	Metadata     string    `gorm:"type:jsonb"` // Дополнительные данные
}

type OrderProcessingResult struct {
	OrderID string `json:"order_id"`
	Action  string `json:"action"` // approved, failed, already_processed
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type AutomaticPaymentResult struct {
	Action  string                  `json:"action"`
	Message string                  `json:"message"`
	Results []OrderProcessingResult `json:"results,omitempty"`
}