package models

import "time"

type TrafficModel struct {
	ID 					string 	`gorm:"primaryKey;type:uuid"`
	MerchantID 			string	`gorm:"index:idx_merchant_trader"`
	TraderID 			string	`gorm:"type:uuid;index:idx_merchant_trader"`
	TraderRewardPercent float64
	PlatformFee			float64 
	TraderPriority 		float64
	Enabled 			bool
	Name				string

	// Поля для антифрода
	AntifraudUnlocked     bool                   `gorm:"default:true"`
	AntifraudLockedAt     *time.Time             
	AntifraudUnlockedAt   *time.Time             
	AntifraudLockReason   string

	// Новые поля для грейс-периода
	ManualUnlockBy        string                 `gorm:"type:uuid"` // ID админа, который разблокировал
	ManualUnlockAt        *time.Time             
	ManualUnlockReason    string                 // Причина разблокировки от админа
	GracePeriodUntil      *time.Time             // До какого времени действует грейс-период

	// Снепшоты состояния на момент разблокировки
	UnlockSnapshot        map[string]interface{} `gorm:"type:jsonb"` // Сохраняем метрики на момент разблокировки

	// Гибкие настройки
	MerchantUnlocked	bool	`gorm:"default:true"`
	TraderUnlocked		bool
	ManuallyUnlocked	bool

	AntifraudRequired bool	

	MerchantDealsDuration time.Duration

	CreatedAt 			time.Time
	UpdatedAt 			time.Time
}