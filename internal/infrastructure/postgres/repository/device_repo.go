package repository

import (
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/mappers"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"gorm.io/gorm"
)

type DefaultDeviceRepository struct {
	DB *gorm.DB
}

func NewDefaultDeviceRepository(db *gorm.DB) *DefaultDeviceRepository {
	return &DefaultDeviceRepository{
		DB: db,
	}
}

func (r *DefaultDeviceRepository) CreateDevice(device *domain.Device) error {
	deviceModel := mappers.ToGORMDevice(device)
	return r.DB.Create(deviceModel).Error
}

func (r *DefaultDeviceRepository) GetTraderDevices(traderID string) ([]*domain.Device, error) {
	var deviceModels []*models.DeviceModel
	if err := r.DB.Model(&models.DeviceModel{}).Where("trader_id = ?", traderID).Find(&deviceModels).Error; err != nil {
		return nil, err
	}

	devices := make([]*domain.Device, len(deviceModels))
	for i, deviceModel := range deviceModels {
		devices[i] = mappers.ToDomainDevice(deviceModel)
	}

	return devices, nil
}

func (r *DefaultDeviceRepository) DeleteDevice(deviceID string) error {
	return r.DB.Delete(&models.DeviceModel{ID: deviceID}).Error
}

func (r *DefaultDeviceRepository) UpdateDevice(deviceID string, params domain.UpdateDeviceParams) error {
	return r.DB.Model(&models.DeviceModel{}).Where("id = ?", deviceID).Updates(map[string]interface{}{
		"enabled": params.Enabled,
		"name": params.Name,
	}).Error
}

func (r *DefaultDeviceRepository) UpdateDeviceLiveness(deviceID string, pingTime time.Time) error {
    return r.DB.Model(&models.DeviceModel{}).
        Where("id = ?", deviceID).
        Updates(map[string]interface{}{
            "device_online":  true,
            "last_ping_at":   pingTime,
            "last_online_at": pingTime,
        }).Error
}

func (r *DefaultDeviceRepository) MarkDevicesOffline(threshold time.Time) error {
    return r.DB.Model(&models.DeviceModel{}).
        Where("device_online = ?", true).
        Where("last_ping_at < ?", threshold).
        Update("device_online", false).Error
}

func (r *DefaultDeviceRepository) GetDeviceByID(deviceID string) (*domain.Device, error) {
    var device models.DeviceModel
    
    err := r.DB.Where("id = ?", deviceID).First(&device).Error
    if err != nil {
        return nil, err
    }
    
    return mappers.ToDomainDevice(&device), nil
}
