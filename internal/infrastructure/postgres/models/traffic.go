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

	// Гибкие настройки
	MerchantEnabled		bool
	TraderEnabled		bool
	DeviceEnabled		bool
	FraudEnabled		bool

	// Для антифрода
	Locked				bool
	LockedAt			time.Time
	UnlockedAt			time.Time		

	CreatedAt 			time.Time
	UpdatedAt 			time.Time
}