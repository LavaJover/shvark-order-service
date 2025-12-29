package usecase

import (
	"fmt"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
)

type TrafficUsecase interface {
	AddTraffic(traffic *domain.Traffic) error
	EditTraffic(input *trafficdto.EditTrafficInput) error
	GetTrafficRecords(page, limit int32) ([]*domain.Traffic, error)
	GetTrafficByID(trafficID string) (*domain.Traffic, error)
	DeleteTraffic(trafficID string) error
	GetTrafficByTraderMerchant(traderID, merchantID string) (*domain.Traffic, error)
	DisableTraderTraffic(traderID string) error
	EnableTraderTraffic(traderID string) error
	GetTraderTrafficStatus(traderID string) (bool, error)
	SetTraderLockTrafficStatus(traderID string, unlocked bool) error
	SetMerchantLockTrafficStatus(traderID string, unlocked bool) error
	SetManuallyLockTrafficStatus(trafficID string, unlocked bool) error
	SetAntifraudLockTrafficStatus(traderID string, unlocked bool) error
	IsTrafficUnlocked(trafficID string) (*trafficdto.TrafficUnlockedResponse, error)
	GetLockStatuses(trafficID string) (*trafficdto.LockStatusesResponse, error)
	GetTrafficByTraderID(traderID string) ([]*domain.Traffic, error) // НОВОЕ
	GetTrafficByMerchantID(merchantID string) ([]*domain.Traffic, error)

	IsTrafficActive(traderID, storeID string) (bool, error)
	GetTrafficByTraderStore(traderID, storeID string) (*domain.Traffic, error)
	GetTrafficByStoreID(storeID string) ([]*domain.Traffic, error)
	GetTrafficWithStoreByTraderStore(traderID, storeID string) (*domain.TrafficWithStore, error)
    ChangeTrafficStore(trafficID, storeID string) error
}

type DefaultTrafficUsecase struct {
	TrafficRepo 	domain.TrafficRepository
	MerchantStoreUC	MerchantStoreUsecase
}

func NewDefaultTrafficUsecase(
	trafficRepo domain.TrafficRepository,
	merchantStoreUC MerchantStoreUsecase,
) *DefaultTrafficUsecase {
	return &DefaultTrafficUsecase{
		TrafficRepo: trafficRepo,
		MerchantStoreUC: merchantStoreUC,
	}
}

func (uc *DefaultTrafficUsecase) AddTraffic(traffic *domain.Traffic) error {
    // Валидация MerchantStore
    if traffic.MerchantStoreID == "" {
        return fmt.Errorf("merchant_store_id is required")
    }
    
    store, err := uc.MerchantStoreUC.ValidateStoreForTraffic(traffic.MerchantStoreID)
    if err != nil {
        return fmt.Errorf("invalid merchant store: %w", err)
    }
    
    // Денормализуем данные из стора
    traffic.MerchantID = store.MerchantID
    
    // Заполняем бизнес-параметры
    traffic.BusinessParams = domain.TrafficBusinessParams{
        StoreName:     store.Name,
        StoreCategory: store.Category,
        MaxDailyDeals: store.MaxDailyDeals,
        MinDealAmount: store.MinDealAmount,
        MaxDealAmount: store.MaxDealAmount,
        Currency:      store.Currency,
    }
    
    return uc.TrafficRepo.CreateTraffic(traffic)
}

// Добавляем недостающий метод EditTraffic
func (uc *DefaultTrafficUsecase) EditTraffic(input *trafficdto.EditTrafficInput) error {
    if input.ID == "" {
        return fmt.Errorf("id is required")
    }
    
    return uc.TrafficRepo.UpdateTraffic(input)
}

// Добавим метод для смены стора в трафике
func (uc *DefaultTrafficUsecase) ChangeTrafficStore(trafficID, storeID string) error {
    if trafficID == "" || storeID == "" {
        return fmt.Errorf("traffic_id and store_id are required")
    }
    
    // Валидация нового стора
    store, err := uc.MerchantStoreUC.ValidateStoreForTraffic(storeID)
    if err != nil {
        return fmt.Errorf("invalid merchant store: %w", err)
    }
    
    // Получаем текущий трафик
    traffic, err := uc.TrafficRepo.GetTrafficByID(trafficID)
    if err != nil {
        return fmt.Errorf("failed to get traffic: %w", err)
    }
    
    if traffic == nil {
        return fmt.Errorf("traffic not found")
    }
    
    // Обновляем ссылку на стор
    traffic.MerchantStoreID = storeID
    traffic.MerchantID = store.MerchantID
    
    traffic.BusinessParams = domain.TrafficBusinessParams{
        StoreName:     store.Name,
        StoreCategory: store.Category,
        MaxDailyDeals: store.MaxDailyDeals,
        MinDealAmount: store.MinDealAmount,
        MaxDealAmount: store.MaxDealAmount,
        Currency:      store.Currency,
    }
    
    // Создаем DTO для обновления
    input := &trafficdto.EditTrafficInput{
        ID: trafficID,
		StoreID: &storeID,
        PlatformFee: &store.PlatformFee,
        // Можно добавить другие поля при необходимости
    }
    
    return uc.TrafficRepo.UpdateTraffic(input)
}

func (uc *DefaultTrafficUsecase) DeleteTraffic(trafficID string) error {
	return uc.TrafficRepo.DeleteTraffic(trafficID)
}

func (uc *DefaultTrafficUsecase) GetTrafficByID(trafficID string) (*domain.Traffic, error) {
	return uc.TrafficRepo.GetTrafficByID(trafficID)
}

func (uc *DefaultTrafficUsecase) GetTrafficRecords(page, limit int32) ([]*domain.Traffic, error) {
	return uc.TrafficRepo.GetTrafficRecords(page, limit)
}

func (uc *DefaultTrafficUsecase) GetTrafficByTraderMerchant(traderID, merchantID string) (*domain.Traffic, error) {
	return uc.TrafficRepo.GetTrafficByTraderMerchant(traderID, merchantID)
}

func (uc *DefaultTrafficUsecase) DisableTraderTraffic(traderID string) error {
	return uc.TrafficRepo.DisableTraderTraffic(traderID)
}

func (uc *DefaultTrafficUsecase) EnableTraderTraffic(traderID string) error {
	return uc.TrafficRepo.EnableTraderTraffic(traderID)
}

func (uc *DefaultTrafficUsecase) GetTraderTrafficStatus(traderID string) (bool, error) {
	return uc.TrafficRepo.GetTraderTrafficStatus(traderID)
}

func (uc *DefaultTrafficUsecase) SetTraderLockTrafficStatus(traderID string, unlocked bool) error {
	return uc.TrafficRepo.SetTraderLockTrafficStatus(traderID, unlocked)
}
func (uc *DefaultTrafficUsecase) SetMerchantLockTrafficStatus(merchantID string, unlocked bool) error {
	return uc.TrafficRepo.SetMerchantLockTrafficStatus(merchantID, unlocked)
}
func (uc *DefaultTrafficUsecase) SetManuallyLockTrafficStatus(trafficID string, unlocked bool) error {
	return uc.TrafficRepo.SetManuallyLockTrafficStatus(trafficID, unlocked)
}
func (uc *DefaultTrafficUsecase) SetAntifraudLockTrafficStatus(traderID string, unlocked bool) error {
	return uc.TrafficRepo.SetAntifraudLockTrafficStatus(traderID, unlocked)
}

// GetLockStatuses возвращает все статусы блокировки для указанного трафика
func (uc *DefaultTrafficUsecase) GetLockStatuses(trafficID string) (*trafficdto.LockStatusesResponse, error) {
	if trafficID == "" {
		return nil, fmt.Errorf("trafficID cannot be empty")
	}

	statuses, err := uc.TrafficRepo.GetLockStatuses(trafficID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lock statuses: %w", err)
	}

	return &trafficdto.LockStatusesResponse{
		TrafficID:         trafficID,
		MerchantUnlocked:  statuses.MerchantUnlocked,
		TraderUnlocked:    statuses.TraderUnlocked,
		AntifraudUnlocked: statuses.AntifraudUnlocked,
		ManuallyUnlocked:  statuses.ManuallyUnlocked,
	}, nil
}

// IsTrafficUnlocked проверяет, разблокирован ли трафик хотя бы одним способом
func (uc *DefaultTrafficUsecase) IsTrafficUnlocked(trafficID string) (*trafficdto.TrafficUnlockedResponse, error) {
	if trafficID == "" {
		return nil, fmt.Errorf("trafficID cannot be empty")
	}

	unlocked, err := uc.TrafficRepo.IsTrafficUnlocked(trafficID)
	if err != nil {
		return nil, fmt.Errorf("failed to check traffic unlock status: %w", err)
	}

	return &trafficdto.TrafficUnlockedResponse{
		TrafficID: trafficID,
		Unlocked:  unlocked,
	}, nil
}

// GetTrafficByTraderID получает все записи трафика для трейдера
func (uc *DefaultTrafficUsecase) GetTrafficByTraderID(traderID string) ([]*domain.Traffic, error) {
    if traderID == "" {
        return nil, fmt.Errorf("trader_id is required")
    }

    return uc.TrafficRepo.GetTrafficByTraderID(traderID)
}

// Обновленный метод для получения трафика по мерчанту (теперь агрегирует по всем сторам)
func (uc *DefaultTrafficUsecase) GetTrafficByMerchantID(merchantID string) ([]*domain.Traffic, error) {
    if merchantID == "" {
        return nil, fmt.Errorf("merchant_id is required")
    }
    
    return uc.TrafficRepo.GetTrafficByMerchantID(merchantID)
}

// Новый метод для получения трафика по стор ID
func (uc *DefaultTrafficUsecase) GetTrafficByStoreID(storeID string) ([]*domain.Traffic, error) {
    if storeID == "" {
        return nil, fmt.Errorf("store_id is required")
    }
    
    return uc.TrafficRepo.GetTrafficByStoreID(storeID)
}

// IsTrafficActive проверяет, активен ли трафик для трейдера и стора
func (uc *DefaultTrafficUsecase) IsTrafficActive(traderID, storeID string) (bool, error) {
    if traderID == "" || storeID == "" {
        return false, fmt.Errorf("trader_id and store_id are required")
    }
    
    traffic, err := uc.GetTrafficByTraderStore(traderID, storeID)
    if err != nil {
        return false, fmt.Errorf("failed to get traffic: %w", err)
    }
    
    if traffic == nil {
        return false, nil
    }
    
    // Проверяем все условия активности
    isActive := traffic.Enabled &&  
                traffic.ActivityParams.TraderUnlocked && 
                traffic.ActivityParams.AntifraudUnlocked && 
                traffic.ActivityParams.ManuallyUnlocked
    
    return isActive, nil
}

// GetTrafficWithStoreByTraderStore возвращает Traffic с данными стора
func (uc *DefaultTrafficUsecase) GetTrafficWithStoreByTraderStore(traderID, storeID string) (*domain.TrafficWithStore, error) {
    if traderID == "" || storeID == "" {
        return nil, fmt.Errorf("trader_id and store_id are required")
    }
    
    return uc.TrafficRepo.GetTrafficWithStoreByTraderStore(traderID, storeID)
}

// GetTrafficByTraderStore возвращает только Traffic
func (uc *DefaultTrafficUsecase) GetTrafficByTraderStore(traderID, storeID string) (*domain.Traffic, error) {
    if traderID == "" || storeID == "" {
        return nil, fmt.Errorf("trader_id and store_id are required")
    }
    
    return uc.TrafficRepo.GetTrafficByTraderStore(traderID, storeID)
}

// IsTrafficActive проверяет, активен ли трафик
// func (uc *DefaultTrafficUsecase) IsTrafficActive(traderID, storeID string) (bool, error) {
//     if traderID == "" || storeID == "" {
//         return false, fmt.Errorf("trader_id and store_id are required")
//     }
    
//     return uc.TrafficRepo.IsTrafficActive(traderID, storeID)
// }