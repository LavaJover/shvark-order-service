package repository

import (
	"fmt"
	"log"
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
    log.Printf("🔄 [REPO] Updating device liveness: deviceID=%s, time=%v", deviceID, pingTime)
    
    result := r.DB.Model(&models.DeviceModel{}).
        Where("id = ?", deviceID).
        Updates(map[string]interface{}{
            "device_online":  true,
            "last_ping_at":   pingTime,
            "last_online_at": pingTime,
        })
    
    if result.Error != nil {
        log.Printf("❌ [REPO] Error updating device liveness: %v", result.Error)
        return result.Error
    }
    
    if result.RowsAffected == 0 {
        log.Printf("⚠️ [REPO] No device found with ID: %s", deviceID)
        return fmt.Errorf("device not found: %s", deviceID)
    }
    
    log.Printf("✅ [REPO] Successfully updated device liveness: deviceID=%s, rows=%d", 
        deviceID, result.RowsAffected)
    return nil
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
    
    // Добавьте логику проверки, действительно ли устройство онлайн
    isOnline := device.DeviceOnline
    if device.LastPingAt != nil {
        // Проверяем, не устарел ли последний пинг
        timeSinceLastPing := time.Since(*device.LastPingAt)
        if timeSinceLastPing > 2*time.Minute {
            isOnline = false
        }
    }
    
    domainDevice := mappers.ToDomainDevice(&device)
    domainDevice.DeviceOnline = isOnline // переопределяем статус
    
    return domainDevice, nil
}
