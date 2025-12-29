package models

import (
    "time"
    
    "gorm.io/gorm"
)

type MerchantStoreModel struct {
    ID            string         `gorm:"primaryKey;type:uuid"`
    MerchantID    string         `gorm:"index:idx_merchant_stores"`
    Name          string         `gorm:"not null"`
    PlatformFee   float64        `gorm:"not null;default:0"`
    IsActive      bool           `gorm:"default:true"`
    DealsDuration time.Duration  `gorm:"not null;default:86400000000000"` // 24 часа по умолчанию
    Description   string
    Category      string         `gorm:"index:idx_store_category"`
    MaxDailyDeals int            `gorm:"default:100"`
    MinDealAmount float64        `gorm:"default:0"`
    MaxDealAmount float64        `gorm:"default:1000000"`
    Currency      string         `gorm:"default:'RUB'"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (MerchantStoreModel) TableName() string {
    return "merchant_stores"
}