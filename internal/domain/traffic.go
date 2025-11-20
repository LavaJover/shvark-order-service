package domain

import (
	"encoding/json"
	"time"

	trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
)

type Traffic struct {
	ID 					string
	MerchantID 			string
	TraderID 			string
	TraderRewardPercent float64
	PlatformFee			float64
	TraderPriority 		float64
	Enabled 			bool // для админов
	Name 				string

	// Новые поля для конфигурации курсов
	ExchangeConfig *ExchangeConfig


	// Гибкие параметры
	ActivityParams 		TrafficActivityParams

	// Для антифрода
	AntifraudParams		TrafficAntifraudParams

	// Бизнес-параметры
	BusinessParams		TrafficBusinessParams
}

// ExchangeConfig - конфигурация биржи и расчета курса
type ExchangeConfig struct {
    // Источник курса
    ExchangeProvider string `json:"exchange_provider"` // "rapira", "bybit", "binance", "manual"
    
    // Настройки позиций в стакане
    OrderBookPositions *OrderBookRange `json:"order_book_positions,omitempty"`
    
    // Искусственная наценка/скидка (в процентах)
    MarkupPercent float64 `json:"markup_percent"`
    
    // Fallback провайдеры на случай ошибок
    FallbackProviders []string `json:"fallback_providers"`
    
    // Минимальный объем для расчета (если применимо)
    MinVolume float64 `json:"min_volume,omitempty"`
    
    // Валютная пара
    CurrencyPair string `json:"currency_pair"` // "USDT/RUB", "BTC/USDT" и т.д.
}

// OrderBookRange - диапазон позиций в стакане
type OrderBookRange struct {
    Start int `json:"start"` // начальная позиция (включительно)
    End   int `json:"end"`   // конечная позиция (включительно)
}

// Методы для сериализации/десериализации JSON
func (t *Traffic) SetExchangeConfig(config *ExchangeConfig) error {
    _, err := json.Marshal(config)
    if err != nil {
        return err
    }
    // Будем хранить как JSON в БД
    return nil
}

func (t *Traffic) GetExchangeConfig() (*ExchangeConfig, error) {
    if t.ExchangeConfig == nil {
        return t.getDefaultExchangeConfig(), nil
    }
    return t.ExchangeConfig, nil
}

func (t *Traffic) getDefaultExchangeConfig() *ExchangeConfig {
    return &ExchangeConfig{
        ExchangeProvider: "rapira",
        OrderBookPositions: &OrderBookRange{Start: 0, End: 4},
        MarkupPercent: 0.0,
        FallbackProviders: []string{"bybit", "binance"},
        CurrencyPair: "USDT/RUB",
    }
}

type TrafficActivityParams struct {
	MerchantUnlocked	bool
	TraderUnlocked		bool
	AntifraudUnlocked	bool
	ManuallyUnlocked	bool
}

type TrafficAntifraudParams struct {
	AntifraudRequired bool
}

type TrafficLockDetails struct {
	LockedAt			time.Time
	UnlockedAt			time.Time
	Reason				string
}

type TrafficBusinessParams struct {
	MerchantDealsDuration time.Duration
}

type TrafficRepository interface {
	CreateTraffic(traffic *Traffic) error
	UpdateTraffic(input *trafficdto.EditTrafficInput) error
	GetTrafficRecords(page, limit int32) ([]*Traffic, error)
	GetTrafficByID(trafficID string) (*Traffic, error)
	DeleteTraffic(trafficID string) error
	GetTrafficByTraderMerchant(traderID, merchantID string) (*Traffic, error)
	DisableTraderTraffic(traderID string) error
	EnableTraderTraffic(traderID string) error
	GetTraderTrafficStatus(traderID string) (bool, error)
	SetTraderLockTrafficStatus(traderID string, unlocked bool) error
	SetMerchantLockTrafficStatus(traderID string, unlocked bool) error
	SetManuallyLockTrafficStatus(trafficID string, unlocked bool) error
	SetAntifraudLockTrafficStatus(traderID string, unlocked bool) error
	IsTrafficUnlocked(trafficID string) (bool, error)
	GetLockStatuses(trafficID string) (*struct {
		MerchantUnlocked  bool
		TraderUnlocked    bool
		AntifraudUnlocked bool
		ManuallyUnlocked  bool
	}, error)
	GetTrafficByTraderID(traderID string) ([]*Traffic, error) // НОВОЕ
	UpdateExchangeConfig(trafficID string, config *ExchangeConfig) error
}