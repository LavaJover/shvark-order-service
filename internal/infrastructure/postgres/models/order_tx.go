package models

import "time"

// OrderTransactionStateModel - модель состояния транзакции операции в БД
type OrderTransactionStateModel struct {
    ID              uint       `gorm:"primaryKey"`
    OrderID         string     `gorm:"index;not null"`
    Operation       string     `gorm:"not null"` // "create", "approve", "cancel"
    StatusChanged   bool       `gorm:"default:false"`
    WalletProcessed bool       `gorm:"default:false"`
    EventPublished  bool       `gorm:"default:false"`
    CallbackSent    bool       `gorm:"default:false"`
    CreatedAt       time.Time  `gorm:"autoCreateTime"`
    CompletedAt     *time.Time `gorm:"default:null"`
    UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
}

func (OrderTransactionStateModel) TableName() string {
    return "order_transaction_states"
}