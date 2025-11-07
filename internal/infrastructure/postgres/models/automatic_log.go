package models

import (
    "time"
)

type AutomaticLogModel struct {
    ID              string    `gorm:"primaryKey;type:uuid"`
    DeviceID        string    `gorm:"index:idx_automatic_logs_device"`
    TraderID        *string    `gorm:"index:idx_automatic_logs_trader"`
    OrderID         *string    `gorm:"type:uuid;index:idx_automatic_logs_order"`
    
    // Данные из уведомления
    Amount          float64
    PaymentSystem   string
    Direction       string
    Methods         string    `gorm:"type:text"` // JSON массив
    ReceivedAt      time.Time
    Text            string    `gorm:"type:text"`
    
    // Результат обработки
    Action          string    // found, not_found, approved, failed, duplicate
    Success         bool
    OrdersFound     int       // Количество найденных заказов
    ErrorMessage    string    `gorm:"type:text"`
    
    // Метаданные
    ProcessingTime  int64     // Время обработки в миллисекундах
    BankName        string
    CardNumber      string
    
    CreatedAt       time.Time `gorm:"index:idx_automatic_logs_created"`
}

func (AutomaticLogModel) TableName() string {
    return "automatic_logs"
}
