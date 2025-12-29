package domain

import (
    "time"
    
    trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
)

type Traffic struct {
    ID                  string
    MerchantStoreID     string    // Теперь ссылаемся на MerchantStore вместо MerchantID
    TraderID            string
    MerchantID          string    // Оставляем для обратной совместимости и быстрого доступа
    
    
    TraderRewardPercent float64
    TraderPriority      float64
    Enabled             bool
    
    // Гибкие параметры
    ActivityParams      TrafficActivityParams
    AntifraudParams     TrafficAntifraudParams
    BusinessParams      TrafficBusinessParams

	CreatedAt time.Time
	UpdatedAt time.Time
}

type TrafficActivityParams struct {
    TraderUnlocked      bool
    AntifraudUnlocked   bool
    ManuallyUnlocked    bool
}

type TrafficAntifraudParams struct {
    AntifraudRequired   bool
    AntifraudLockedAt   *time.Time
    AntifraudUnlockedAt *time.Time
    AntifraudLockReason string
}

type TrafficLockDetails struct {
    LockedAt            time.Time
    UnlockedAt          time.Time
    Reason              string
}

type TrafficBusinessParams struct {
    // Денормализованные поля из MerchantStore
    StoreName          string
    StoreCategory      string
    MaxDailyDeals      int
    MinDealAmount      float64
    MaxDealAmount      float64
    Currency           string
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
    GetTrafficByTraderID(traderID string) ([]*Traffic, error)
    GetTrafficByMerchantID(merchantID string) ([]*Traffic, error)
	GetTrafficWithStoreByTraderStore(traderID, storeID string) (*TrafficWithStore, error)
	GetTrafficByTraderStore(traderID, storeID string) (*Traffic, error)
	GetActiveTrafficsByStore(storeID string) ([]*Traffic, error)
	GetTrafficByStoreID(storeID string) ([]*Traffic, error)
	IsTrafficActive(traderID, storeID string) (bool, error)
	UpdateTrafficStore(trafficID, storeID, merchantID string) error // Добавляем этот метод
}

// TrafficWithStore для JOIN запросов
type TrafficWithStore struct {
    Traffic
    Store MerchantStore
}