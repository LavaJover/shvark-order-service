package repository

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
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
	}

	if err := r.DB.Create(&trafficModel).Error; err != nil {
		return err
	}

	traffic.ID = trafficModel.ID
	return nil
}

func (r *DefaultTrafficRepository) UpdateTraffic(traffic *domain.Traffic) error {
    // Используем карту для явного указания полей к обновлению
    updateData := map[string]interface{}{
        "trader_reward_percent": traffic.TraderRewardPercent,
        "trader_priority":      traffic.TraderPriority,
        "enabled":              traffic.Enabled,
        "platform_fee":         traffic.PlatformFee,
    }

    // Обновляем запись с явным указанием полей
    if err := r.DB.Model(&models.TrafficModel{}).
        Where("id = ?", traffic.ID).
        Updates(updateData).Error; err != nil {
        return err
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