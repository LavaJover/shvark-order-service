package mappers

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
)

func ToGORMDevice(device *domain.Device) *models.DeviceModel {
	return &models.DeviceModel{
		ID: device.DeviceID,
		Name: device.DeviceName,
		TraderID: device.TraderID,
		Enabled: device.Enabled,
		DeviceOnline: device.DeviceOnline,
		LastPingAt: device.LastPingAt,
		LastOnlineAt: device.LastOnlineAt,
		CreatedAt: device.CreatedAt,
		UpdatedAt: device.UpdatedAt,
	}
}

func ToDomainDevice(model *models.DeviceModel) *domain.Device {
	return &domain.Device{
		DeviceID: model.ID,
		DeviceName: model.Name,
		TraderID: model.TraderID,
		Enabled: model.Enabled,
		DeviceOnline: model.DeviceOnline,
		LastPingAt: model.LastPingAt,
		LastOnlineAt: model.LastOnlineAt,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}