package logger

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type OrderCreatedEvent struct {
	ID 				uint 	`gorm:"primaryKey"`
	RequestID 		string
	OrderID 		string
	MerchantID 		string
	TraderID 		string
	AmountFiat 		float64
	Currency 		string
	BankName 		string
	BankCode 		string
	PaymentSystem 	string
	Phone 			string
	CardNumber		string
	Timestamp 		time.Time
}

type OrderFailedEvent struct {
	ID 				uint `gorm:"primaryKey"`
	RequestID 		string
	MerchantID 		string
	Reason 			string
	AmountFiat 		float64
	Currency 		string
	BankCode 		string
	PaymentSystem 	string
	Timestamp 		time.Time
}

type OrderEventLogger interface {
	LogOrderCreated(ctx context.Context, event OrderCreatedEvent) error
	LogOrderFailed(ctx context.Context, event OrderFailedEvent) error
}

type PGOrderEventLogger struct {
	db *gorm.DB
}

func NewPGOrderEventLogger(db *gorm.DB) *PGOrderEventLogger {
	return &PGOrderEventLogger{db: db}
}

func (l *PGOrderEventLogger) LogOrderCreated(ctx context.Context, event OrderCreatedEvent) error {
	return l.db.WithContext(ctx).Create(&event).Error
}

func (l *PGOrderEventLogger) LogOrderFailed(ctx context.Context, event OrderFailedEvent) error {
	return l.db.WithContext(ctx).Create(&event).Error
}