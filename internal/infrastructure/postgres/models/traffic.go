package models

import "time"

type TrafficModel struct {
    ID                    string                 `gorm:"primaryKey;type:uuid"`
    MerchantStoreID       string                 `gorm:"type:uuid;index:idx_traffic_store"`
    MerchantID            string                 `gorm:"index:idx_merchant_trader"`
    TraderID              string                 `gorm:"type:uuid;index:idx_merchant_trader"`

    
    TraderRewardPercent   float64
    TraderPriority        float64
    Enabled               bool                   `gorm:"default:true"`
    Name                  string
    
    // Бизнес параметры из стора
    StoreName             string
    StoreCategory         string
    MaxDailyDeals         int
    MinDealAmount         float64
    MaxDealAmount         float64
    Currency              string
    
    // Поля для антифрода
    AntifraudUnlocked     bool                   `gorm:"default:true"`
    AntifraudLockedAt     *time.Time
    AntifraudUnlockedAt   *time.Time
    AntifraudLockReason   string
    
    // Новые поля для грейс-периода
    ManualUnlockBy        string
    ManualUnlockAt        *time.Time
    ManualUnlockReason    string
    GracePeriodUntil      *time.Time
    
    // Снепшоты состояния на момент разблокировки
    UnlockSnapshot        map[string]interface{} `gorm:"type:jsonb"`
    
    // Гибкие настройки
    TraderUnlocked        bool                   `gorm:"default:true"`
    ManuallyUnlocked      bool                   `gorm:"default:false"`
    
    AntifraudRequired     bool                   `gorm:"default:false"`
    
    CreatedAt             time.Time
    UpdatedAt             time.Time
}

func (TrafficModel) TableName() string {
    return "traffics"
}