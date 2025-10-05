package domain

import "time"

type Device struct {
	DeviceID 	string
	DeviceName 	string
	TraderID 	string
	Enabled 	bool

	DeviceOnline bool

	CreatedAt    time.Time
	UpdatedAt	 time.Time
}

type DeviceRepository interface {
	CreateDevice(device *Device) error
	GetTraderDevices(traderID string) ([]*Device, error)
	DeleteDevice(deviceID string) error
	UpdateDevice(deviceID string, params UpdateDeviceParams) error
}

type UpdateDeviceParams struct {
	Name 	string
	Enabled bool
}