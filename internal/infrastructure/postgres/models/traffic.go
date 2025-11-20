package models

import (
	"encoding/json"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
)

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
	ManualUnlockBy        string                 // ID админа, который разблокировал
	ManualUnlockAt        *time.Time             
	ManualUnlockReason    string                 // Причина разблокировки от админа
	GracePeriodUntil      *time.Time             // До какого времени действует грейс-период

	// Снепшоты состояния на момент разблокировки
	UnlockSnapshot        map[string]interface{} `gorm:"type:jsonb"` // Сохраняем метрики на момент разблокировки

	// Новые поля для конфигурации курсов
    ExchangeConfigJSON   string    `gorm:"type:jsonb"`

	// Гибкие настройки
	MerchantUnlocked	bool	`gorm:"default:true"`
	TraderUnlocked		bool
	ManuallyUnlocked	bool

	AntifraudRequired bool	

	MerchantDealsDuration time.Duration

	CreatedAt 			time.Time
	UpdatedAt 			time.Time
}

// GetExchangeConfig возвращает распарсенную конфигурацию
func (t *TrafficModel) GetExchangeConfig() (*domain.ExchangeConfig, error) {
    if t.ExchangeConfigJSON == "" {
        return t.getDefaultConfig(), nil
    }
    
    var config domain.ExchangeConfig
    if err := json.Unmarshal([]byte(t.ExchangeConfigJSON), &config); err != nil {
        return t.getDefaultConfig(), nil
    }
    
    return &config, nil
}

// SetExchangeConfig устанавливает конфигурацию
func (t *TrafficModel) SetExchangeConfig(config *domain.ExchangeConfig) error {
    data, err := json.Marshal(config)
    if err != nil {
        return err
    }
    t.ExchangeConfigJSON = string(data)
    return nil
}

func (t *TrafficModel) getDefaultConfig() *domain.ExchangeConfig {
    return &domain.ExchangeConfig{
        ExchangeProvider: "rapira",
        OrderBookPositions: &domain.OrderBookRange{Start: 0, End: 4},
        MarkupPercent: 0.0,
        FallbackProviders: []string{"bybit", "binance"},
        CurrencyPair: "USDT/RUB",
    }
}