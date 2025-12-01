package repository

import (
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DefaultTrafficRepository struct {
	DB *gorm.DB
}

func NewDefaultTrafficRepository(db *gorm.DB) *DefaultTrafficRepository {
	return &DefaultTrafficRepository{DB: db}
}

func (r *DefaultTrafficRepository) CreateTraffic(traffic *domain.Traffic) error {
	trafficModel := models.TrafficModel{
		ID: uuid.New().String(),
		MerchantID: traffic.MerchantID,
		TraderID: traffic.TraderID,
		TraderRewardPercent: traffic.TraderRewardPercent,
		TraderPriority: traffic.TraderPriority,
		Enabled: traffic.Enabled,
		PlatformFee: traffic.PlatformFee,
		MerchantUnlocked: traffic.ActivityParams.MerchantUnlocked,
		TraderUnlocked: traffic.ActivityParams.TraderUnlocked,
		AntifraudUnlocked: traffic.ActivityParams.AntifraudUnlocked,
		ManuallyUnlocked: traffic.ActivityParams.ManuallyUnlocked,
		AntifraudRequired: traffic.AntifraudParams.AntifraudRequired,
		MerchantDealsDuration: traffic.BusinessParams.MerchantDealsDuration,
		Name: traffic.Name,
	}

	if err := r.DB.Create(&trafficModel).Error; err != nil {
		return err
	}

	traffic.ID = trafficModel.ID
	return nil
}

func (r *DefaultTrafficRepository) UpdateTraffic(input *trafficdto.EditTrafficInput) error {
	updates := make(map[string]interface{})
	
	// Обновляем простые поля если они переданы
	if input.MerchantID != nil {
		updates["merchant_id"] = *input.MerchantID
	}
	if input.TraderID != nil {
		updates["trader_id"] = *input.TraderID
	}
	if input.TraderReward != nil {
		updates["trader_reward_percent"] = *input.TraderReward
	}
	if input.TraderPriority != nil {
		updates["trader_priority"] = *input.TraderPriority
	}
	if input.PlatformFee != nil {
		updates["platform_fee"] = *input.PlatformFee
	}
	if input.Enabled != nil {
		updates["enabled"] = *input.Enabled
	}
	if input.Name != nil {
		updates["name"] = *input.Name
	}

	// Обновляем вложенные структуры если они переданы
	if input.ActivityParams != nil {
		updates["merchant_unlocked"] = input.ActivityParams.MerchantUnlocked
		updates["trader_unlocked"] = input.ActivityParams.TraderUnlocked
		updates["antifraud_unlocked"] = input.ActivityParams.AntifraudUnlocked
		updates["manually_unlocked"] = input.ActivityParams.ManuallyUnlocked
	}

	if input.AntifraudParams != nil {
		updates["antifraud_required"] = input.AntifraudParams.AntifraudRequired
	}

	if input.BusinessParams != nil {
		updates["merchant_deals_duration"] = input.BusinessParams.MerchantDealsDuration
	}

	// Добавляем updated_at
	updates["updated_at"] = time.Now()

	// Выполняем обновление
	result := r.DB.Model(&models.TrafficModel{}).
		Where("id = ?", input.ID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("traffic record with id %s not found", input.ID)
	}

	return nil
}

func (r *DefaultTrafficRepository) DeleteTraffic(trafficID string) error {
	if err := r.DB.Delete(&models.TrafficModel{ID: trafficID}).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultTrafficRepository) GetTrafficRecords(page, limit int32) ([]*domain.Traffic, error) {
	var trafficModels []models.TrafficModel
	var total int64

	// Подсчёт числа записей
	r.DB.Model(&models.TrafficModel{}).Count(&total)

	// Параметры пагинации
	offset := (page-1) * limit
	// totalPages := (int32(total) + limit - 1) / limit

	if err := r.DB.Offset(int(offset)).Limit(int(limit)).Order("created_at DESC").Find(&trafficModels).Error; err != nil {
		return nil, err
	}

	trafficRecords := make([]*domain.Traffic, len(trafficModels))
	for i, trafficModel := range trafficModels {
		trafficRecords[i] = &domain.Traffic{
			ID: trafficModel.ID,
			MerchantID: trafficModel.MerchantID,
			TraderID: trafficModel.TraderID,
			TraderRewardPercent: trafficModel.TraderRewardPercent,
			TraderPriority: trafficModel.TraderPriority,
			Enabled: trafficModel.Enabled,
			PlatformFee: trafficModel.PlatformFee,
			Name: trafficModel.Name,
			ActivityParams: domain.TrafficActivityParams{
				MerchantUnlocked: trafficModel.MerchantUnlocked,
				TraderUnlocked: trafficModel.TraderUnlocked,
				AntifraudUnlocked: trafficModel.AntifraudUnlocked,
				ManuallyUnlocked: trafficModel.ManuallyUnlocked,
			},
			AntifraudParams: domain.TrafficAntifraudParams{
				AntifraudRequired: trafficModel.AntifraudRequired,
			},
			BusinessParams: domain.TrafficBusinessParams{
				MerchantDealsDuration: trafficModel.MerchantDealsDuration,
			},
		}
	}

	return trafficRecords, nil
}

func (r *DefaultTrafficRepository) GetTrafficByID(trafficID string) (*domain.Traffic, error) {
	var trafficModel models.TrafficModel
	if err := r.DB.Where("id = ?", trafficID).First(&trafficModel).Error; err != nil {
		return nil, err
	}

	return &domain.Traffic{
		ID: trafficModel.ID,
		MerchantID: trafficModel.MerchantID,
		TraderID: trafficModel.TraderID,
		TraderRewardPercent: trafficModel.TraderRewardPercent,
		TraderPriority: trafficModel.TraderPriority,
		Enabled: trafficModel.Enabled,
		PlatformFee: trafficModel.PlatformFee,
		Name: trafficModel.Name,
		ActivityParams: domain.TrafficActivityParams{
			MerchantUnlocked: trafficModel.MerchantUnlocked,
			TraderUnlocked: trafficModel.TraderUnlocked,
			AntifraudUnlocked: trafficModel.AntifraudUnlocked,
			ManuallyUnlocked: trafficModel.ManuallyUnlocked,
		},
		AntifraudParams: domain.TrafficAntifraudParams{
			AntifraudRequired: trafficModel.AntifraudRequired,
		},
		BusinessParams: domain.TrafficBusinessParams{
			MerchantDealsDuration: trafficModel.MerchantDealsDuration,
		},
	}, nil
}

func (r *DefaultTrafficRepository) GetTrafficByTraderMerchant(traderID, merchantID string) (*domain.Traffic, error) {
	var trafficModel models.TrafficModel
	if err := r.DB.Where("trader_id = ? AND merchant_id = ?", traderID, merchantID).First(&trafficModel).Error; err != nil {
		return nil, err
	}

	return &domain.Traffic{
		ID: trafficModel.ID,
		MerchantID: trafficModel.MerchantID,
		TraderID: trafficModel.TraderID,
		TraderRewardPercent: trafficModel.TraderRewardPercent,
		TraderPriority: trafficModel.TraderPriority,
		Enabled: trafficModel.Enabled,
		PlatformFee: trafficModel.PlatformFee,
		Name: trafficModel.Name,
		ActivityParams: domain.TrafficActivityParams{
			MerchantUnlocked: trafficModel.MerchantUnlocked,
			TraderUnlocked: trafficModel.TraderUnlocked,
			AntifraudUnlocked: trafficModel.AntifraudUnlocked,
			ManuallyUnlocked: trafficModel.ManuallyUnlocked,
		},
		AntifraudParams: domain.TrafficAntifraudParams{
			AntifraudRequired: trafficModel.AntifraudRequired,
		},
		BusinessParams: domain.TrafficBusinessParams{
			MerchantDealsDuration: trafficModel.MerchantDealsDuration,
		},
	}, nil
}

func (r *DefaultTrafficRepository) DisableTraderTraffic(traderID string) error {
	err := r.DB.Model(&models.TrafficModel{}).Where("trader_id = ?", traderID).Update("enabled", false).Error
	return err
}

func (r *DefaultTrafficRepository) EnableTraderTraffic(traderID string) error {
	err := r.DB.Model(&models.TrafficModel{}).Where("trader_id = ?", traderID).Update("enabled", true).Error
	return err
}

func (r *DefaultTrafficRepository) GetTraderTrafficStatus(traderID string) (bool, error) {
	var count int64
	if err := r.DB.Model(&models.TrafficModel{}).Where("trader_id = ? AND enabled = ?", traderID, true).Count(&count).Error; err != nil {
		return false, nil
	}

	return count > 0, nil
}

func (r *DefaultTrafficRepository) SetTraderLockTrafficStatus(traderID string, unlocked bool) error {
	result := r.DB.Model(&models.TrafficModel{}).
		Where("trader_id = ?", traderID).
		Updates(map[string]interface{}{
			"trader_unlocked": unlocked,
			"updated_at":      time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update trader lock status for trader %s: %w", traderID, result.Error)
	}

	return nil
}

func (r *DefaultTrafficRepository) SetMerchantLockTrafficStatus(merchantID string, unlocked bool) error {
	result := r.DB.Model(&models.TrafficModel{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"merchant_unlocked": unlocked,
			"updated_at":        time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update merchant lock status for merchant %s: %w", merchantID, result.Error)
	}

	return nil
}

func (r *DefaultTrafficRepository) SetManuallyLockTrafficStatus(trafficID string, unlocked bool) error {
	result := r.DB.Model(&models.TrafficModel{}).
		Where("id = ?", trafficID).
		Updates(map[string]interface{}{
			"manually_unlocked": unlocked,
			"updated_at":        time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update manual lock status for traffic %s: %w", trafficID, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("traffic record with id %s not found", trafficID)
	}

	return nil
}

func (r *DefaultTrafficRepository) SetAntifraudLockTrafficStatus(traderID string, unlocked bool) error {
	result := r.DB.Model(&models.TrafficModel{}).
		Where("trader_id = ?", traderID).
		Updates(map[string]interface{}{
			"antifraud_unlocked": unlocked,
			"updated_at":         time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update antifraud lock status for trader %s: %w", traderID, result.Error)
	}

	return nil
}

// Блокировка/разблокировка всех статусов для конкретного трафика
func (r *DefaultTrafficRepository) SetAllLockStatuses(trafficID string, statuses struct {
	MerchantUnlocked *bool
	TraderUnlocked   *bool
	AntifraudUnlocked *bool
	ManuallyUnlocked  *bool
}) error {
	updates := make(map[string]interface{})
	updates["updated_at"] = time.Now()

	if statuses.MerchantUnlocked != nil {
		updates["merchant_unlocked"] = *statuses.MerchantUnlocked
	}
	if statuses.TraderUnlocked != nil {
		updates["trader_unlocked"] = *statuses.TraderUnlocked
	}
	if statuses.AntifraudUnlocked != nil {
		updates["antifraud_unlocked"] = *statuses.AntifraudUnlocked
	}
	if statuses.ManuallyUnlocked != nil {
		updates["manually_unlocked"] = *statuses.ManuallyUnlocked
	}

	result := r.DB.Model(&models.TrafficModel{}).
		Where("id = ?", trafficID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update lock statuses for traffic %s: %w", trafficID, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("traffic record with id %s not found", trafficID)
	}

	return nil
}

// Получение текущих статусов блокировки
func (r *DefaultTrafficRepository) GetLockStatuses(trafficID string) (*struct {
	MerchantUnlocked  bool
	TraderUnlocked    bool
	AntifraudUnlocked bool
	ManuallyUnlocked  bool
}, error) {
	var traffic models.TrafficModel
	result := r.DB.Select(
		"merchant_unlocked",
		"trader_unlocked", 
		"antifraud_unlocked",
		"manually_unlocked",
	).First(&traffic, "id = ?", trafficID)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get lock statuses for traffic %s: %w", trafficID, result.Error)
	}

	return &struct {
		MerchantUnlocked  bool
		TraderUnlocked    bool
		AntifraudUnlocked bool
		ManuallyUnlocked  bool
	}{
		MerchantUnlocked:  traffic.MerchantUnlocked,
		TraderUnlocked:    traffic.TraderUnlocked,
		AntifraudUnlocked: traffic.AntifraudUnlocked,
		ManuallyUnlocked:  traffic.ManuallyUnlocked,
	}, nil
}

// Проверка, разблокирован ли трафик (хотя бы одним способом)
func (r *DefaultTrafficRepository) IsTrafficUnlocked(trafficID string) (bool, error) {
	var traffic models.TrafficModel
	result := r.DB.Select(
		"merchant_unlocked",
		"trader_unlocked",
		"antifraud_unlocked", 
		"manually_unlocked",
	).First(&traffic, "id = ?", trafficID)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check traffic unlock status for %s: %w", trafficID, result.Error)
	}

	return traffic.MerchantUnlocked || traffic.TraderUnlocked || 
	       traffic.AntifraudUnlocked || traffic.ManuallyUnlocked, nil
}

// GetTrafficByTraderID получает все записи трафика для трейдера
func (r *DefaultTrafficRepository) GetTrafficByTraderID(traderID string) ([]*domain.Traffic, error) {
    var trafficModels []models.TrafficModel
    
    err := r.DB.Where("trader_id = ?", traderID).Find(&trafficModels).Error
    if err != nil {
        return nil, err
    }

    traffics := make([]*domain.Traffic, 0, len(trafficModels))
    for _, tm := range trafficModels {
        traffics = append(traffics, &domain.Traffic{
			ID: tm.ID,
			MerchantID: tm.MerchantID,
			TraderID: tm.TraderID,
			TraderRewardPercent: tm.TraderRewardPercent,
			TraderPriority: tm.TraderPriority,
			Enabled: tm.Enabled,
			PlatformFee: tm.PlatformFee,
			Name: tm.Name,
			ActivityParams: domain.TrafficActivityParams{
				MerchantUnlocked: tm.MerchantUnlocked,
				TraderUnlocked: tm.TraderUnlocked,
				AntifraudUnlocked: tm.AntifraudUnlocked,
				ManuallyUnlocked: tm.ManuallyUnlocked,
			},
			AntifraudParams: domain.TrafficAntifraudParams{
				AntifraudRequired: tm.AntifraudRequired,
			},
			BusinessParams: domain.TrafficBusinessParams{
				MerchantDealsDuration: tm.MerchantDealsDuration,
			},
		})
    }

    return traffics, nil
}

func (r *DefaultTrafficRepository) GetTrafficByMerchantID(merchantID string) ([]*domain.Traffic, error) {
    var trafficModels []models.TrafficModel
    
    err := r.DB.Where("merchant_id = ?", merchantID).Find(&trafficModels).Error
    if err != nil {
        return nil, err
    }

    traffics := make([]*domain.Traffic, 0, len(trafficModels))
    for _, tm := range trafficModels {
        traffics = append(traffics, &domain.Traffic{
			ID: tm.ID,
			MerchantID: tm.MerchantID,
			TraderID: tm.TraderID,
			TraderRewardPercent: tm.TraderRewardPercent,
			TraderPriority: tm.TraderPriority,
			Enabled: tm.Enabled,
			PlatformFee: tm.PlatformFee,
			Name: tm.Name,
			ActivityParams: domain.TrafficActivityParams{
				MerchantUnlocked: tm.MerchantUnlocked,
				TraderUnlocked: tm.TraderUnlocked,
				AntifraudUnlocked: tm.AntifraudUnlocked,
				ManuallyUnlocked: tm.ManuallyUnlocked,
			},
			AntifraudParams: domain.TrafficAntifraudParams{
				AntifraudRequired: tm.AntifraudRequired,
			},
			BusinessParams: domain.TrafficBusinessParams{
				MerchantDealsDuration: tm.MerchantDealsDuration,
			},
		})
    }

    return traffics, nil
}