package domain

import "time"

type MerchantStore struct {
    ID             string
    MerchantID     string
    PlatformFee    float64
    IsActive       bool
    Name           string
    DealsDuration  time.Duration
    Description    string    // Добавим описание
    Category       string    // Категория стора (например: "exchange", "shop", "service")
    MaxDailyDeals  int       // Максимальное количество сделок в день
    MinDealAmount  float64   // Минимальная сумма сделки
    MaxDealAmount  float64   // Максимальная сумма сделка
    Currency       string    // Основная валюта
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type MerchantStoreRepository interface {
    CreateMerchantStore(store *MerchantStore) error
    UpdateMerchantStore(store *MerchantStore) error
    DeleteMerchantStore(id string) error
    GetMerchantStoreByID(id string) (*MerchantStore, error)
    GetMerchantStores(page, limit int32) ([]*MerchantStore, error)
    GetStoresByMerchantID(merchantID string) ([]*MerchantStore, error)
    GetActiveStoresByMerchantID(merchantID string) ([]*MerchantStore, error)
}