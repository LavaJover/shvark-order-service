package usecase

import (
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	merchantdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/merchant"
	"github.com/google/uuid"
)

type MerchantStoreUsecase interface {
    CreateMerchantStore(input *merchantdto.CreateStoreInput) (*domain.MerchantStore, error)
    UpdateMerchantStore(input *merchantdto.UpdateStoreInput) (*domain.MerchantStore, error)
    DeleteMerchantStore(id string) error
    GetMerchantStoreByID(id string) (*domain.MerchantStore, error)
    GetMerchantStores(page, limit int32) ([]*domain.MerchantStore, error)
    GetStoresByMerchantID(merchantID string) ([]*domain.MerchantStore, error)
    GetActiveStoresByMerchantID(merchantID string) ([]*domain.MerchantStore, error)
    ValidateStoreForTraffic(storeID string) (*domain.MerchantStore, error)
}

type DefaultMerchantStoreUsecase struct {
    StoreRepo domain.MerchantStoreRepository
    TrafficRepo domain.TrafficRepository
}

func NewDefaultMerchantStoreUsecase(
    storeRepo domain.MerchantStoreRepository,
    trafficRepo domain.TrafficRepository,
) *DefaultMerchantStoreUsecase {
    return &DefaultMerchantStoreUsecase{
        StoreRepo: storeRepo,
        TrafficRepo: trafficRepo,
    }
}

func (uc *DefaultMerchantStoreUsecase) CreateMerchantStore(input *merchantdto.CreateStoreInput) (*domain.MerchantStore, error) {
    if input.MerchantID == "" {
        return nil, fmt.Errorf("merchant_id is required")
    }
    
    if input.Name == "" {
        return nil, fmt.Errorf("store name is required")
    }
    
    if input.PlatformFee < 0 || input.PlatformFee > 100 {
        return nil, fmt.Errorf("platform fee must be between 0 and 100")
    }
    
    store := &domain.MerchantStore{
        ID:            uuid.New().String(), // Нужна функция генерации UUID
        MerchantID:    input.MerchantID,
        PlatformFee:   input.PlatformFee,
        IsActive:      true, // По умолчанию активен
        Name:          input.Name,
        DealsDuration: input.DealsDuration,
        Description:   input.Description,
        Category:      input.Category,
        MaxDailyDeals: input.MaxDailyDeals,
        MinDealAmount: input.MinDealAmount,
        MaxDealAmount: input.MaxDealAmount,
        Currency:      input.Currency,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }
    
    if err := uc.StoreRepo.CreateMerchantStore(store); err != nil {
        return nil, fmt.Errorf("failed to create merchant store: %w", err)
    }
    
    return store, nil
}

func (uc *DefaultMerchantStoreUsecase) UpdateMerchantStore(input *merchantdto.UpdateStoreInput) (*domain.MerchantStore, error) {
    if input.ID == "" {
        return nil, fmt.Errorf("store id is required")
    }
    
    store, err := uc.StoreRepo.GetMerchantStoreByID(input.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get store: %w", err)
    }
    
    if store == nil {
        return nil, fmt.Errorf("store not found")
    }
    
    // Обновляем только предоставленные поля
    if input.Name != "" {
        store.Name = input.Name
    }
    
    if input.PlatformFee >= 0 && input.PlatformFee <= 100 {
        store.PlatformFee = input.PlatformFee
    }
    
    if input.DealsDuration > 0 {
        store.DealsDuration = input.DealsDuration
    }
    
    if input.Description != "" {
        store.Description = input.Description
    }
    
    if input.Category != "" {
        store.Category = input.Category
    }
    
    if input.MaxDailyDeals > 0 {
        store.MaxDailyDeals = input.MaxDailyDeals
    }
    
    if input.MinDealAmount >= 0 {
        store.MinDealAmount = input.MinDealAmount
    }
    
    if input.MaxDealAmount > 0 {
        store.MaxDealAmount = input.MaxDealAmount
    }
    
    if input.Currency != "" {
        store.Currency = input.Currency
    }
    
    store.IsActive = input.IsActive
    store.UpdatedAt = time.Now()
    
    if err := uc.StoreRepo.UpdateMerchantStore(store); err != nil {
        return nil, fmt.Errorf("failed to update merchant store: %w", err)
    }
    
    return store, nil
}

func (uc *DefaultMerchantStoreUsecase) DeleteMerchantStore(id string) error {
    // Проверяем, есть ли связанные трафики
    traffics, err := uc.TrafficRepo.GetTrafficByStoreID(id)
    if err != nil {
        return fmt.Errorf("failed to check related traffics: %w", err)
    }
    
    if len(traffics) > 0 {
        return fmt.Errorf("cannot delete store with active traffics")
    }
    
    return uc.StoreRepo.DeleteMerchantStore(id)
}

func (uc *DefaultMerchantStoreUsecase) GetMerchantStoreByID(id string) (*domain.MerchantStore, error) {
    return uc.StoreRepo.GetMerchantStoreByID(id)
}

func (uc *DefaultMerchantStoreUsecase) GetMerchantStores(page, limit int32) ([]*domain.MerchantStore, error) {
    return uc.StoreRepo.GetMerchantStores(page, limit)
}

func (uc *DefaultMerchantStoreUsecase) GetStoresByMerchantID(merchantID string) ([]*domain.MerchantStore, error) {
    return uc.StoreRepo.GetStoresByMerchantID(merchantID)
}

func (uc *DefaultMerchantStoreUsecase) GetActiveStoresByMerchantID(merchantID string) ([]*domain.MerchantStore, error) {
    return uc.StoreRepo.GetActiveStoresByMerchantID(merchantID)
}

func (uc *DefaultMerchantStoreUsecase) ValidateStoreForTraffic(storeID string) (*domain.MerchantStore, error) {
    store, err := uc.StoreRepo.GetMerchantStoreByID(storeID)
    if err != nil {
        return nil, fmt.Errorf("failed to get store: %w", err)
    }
    
    if store == nil {
        return nil, fmt.Errorf("store not found")
    }
    
    if !store.IsActive {
        return nil, fmt.Errorf("store is not active")
    }
    
    return store, nil
}