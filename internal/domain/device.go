package domain

import "time"

type Device struct {
	DeviceID 	string
	DeviceName 	string
	TraderID 	string
	Enabled 	bool

	// Статус онлайн
	DeviceOnline bool
	LastPingAt   *time.Time // Время последнего пинга
	LastOnlineAt *time.Time // Время когда был онлайн последний раз

	CreatedAt    time.Time
	UpdatedAt	 time.Time
}

type DeviceRepository interface {
	CreateDevice(device *Device) error
	GetTraderDevices(traderID string) ([]*Device, error)
	DeleteDevice(deviceID string) error
	UpdateDevice(deviceID string, params UpdateDeviceParams) error
	UpdateDeviceLiveness(deviceID string, pingTime time.Time) error
	MarkDevicesOffline(threshold time.Time) error
	GetDeviceByID(deviceID string) (*Device, error)
}

type UpdateDeviceParams struct {
	Name 	string
	Enabled bool
}