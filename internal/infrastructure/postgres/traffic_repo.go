package postgres

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
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
	trafficModel := TrafficModel{
		ID: uuid.New().String(),
		MerchantID: traffic.MerchantID,
		TraderID: traffic.TraderID,
		TraderRewardPercent: traffic.TraderRewardPercent,
		TraderPriority: traffic.TraderPriority,
		Enabled: traffic.Enabled,
	}

	if err := r.DB.Create(&trafficModel).Error; err != nil {
		return err
	}

	traffic.ID = trafficModel.ID
	return nil
}

func (r *DefaultTrafficRepository) UpdateTraffic(traffic *domain.Traffic) error {
	trafficModel := TrafficModel{
		ID: traffic.ID,
		MerchantID: traffic.MerchantID,
		TraderID: traffic.TraderID,
		TraderRewardPercent: traffic.TraderRewardPercent,
		TraderPriority: traffic.TraderPriority,
		Enabled: traffic.Enabled,
	}

	if err := r.DB.Model(&trafficModel).Updates(map[string]interface{}{
		"enabled": trafficModel.Enabled,
		"trader_reward_percent": trafficModel.TraderRewardPercent,
		"trader_priority": trafficModel.TraderPriority,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultTrafficRepository) DeleteTraffic(trafficID string) error {
	if err := r.DB.Delete(&TrafficModel{ID: trafficID}).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultTrafficRepository) GetTrafficRecords(page, limit int32) ([]*domain.Traffic, error) {
	var trafficModels []TrafficModel
	var total int64

	// Подсчёт числа записей
	r.DB.Model(&TrafficModel{}).Count(&total)

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
		}
	}

	return trafficRecords, nil
}

func (r *DefaultTrafficRepository) GetTrafficByID(trafficID string) (*domain.Traffic, error) {
	var trafficModel TrafficModel
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
	}, nil
}

func (r *DefaultTrafficRepository) GetTrafficByTraderMerchant(traderID, merchantID string) (*domain.Traffic, error) {
	var trafficModel TrafficModel
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
	}, nil
}

func (r *DefaultTrafficRepository) DisableTraderTraffic(traderID string) error {
	err := r.DB.Model(&TrafficModel{}).Where("trader_id = ?", traderID).Update("enabled", false).Error
	return err
}