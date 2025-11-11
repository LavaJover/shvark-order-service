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
}

type DefaultTrafficUsecase struct {
	TrafficRepo domain.TrafficRepository
}

func NewDefaultTrafficUsecase(trafficRepo domain.TrafficRepository) *DefaultTrafficUsecase {
	return &DefaultTrafficUsecase{TrafficRepo: trafficRepo}
}

func (uc *DefaultTrafficUsecase) AddTraffic(traffic *domain.Traffic) error {
	return uc.TrafficRepo.CreateTraffic(traffic)
}

func (uc *DefaultTrafficUsecase) EditTraffic(input *trafficdto.EditTrafficInput) error {
	// Можно добавить валидацию или бизнес-логику перед обновлением
	if input.ID == "" {
		return fmt.Errorf("id is required")
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
