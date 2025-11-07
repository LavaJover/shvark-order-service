package usecase

import (
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	devicedto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/device"
	"github.com/jaevor/go-nanoid"
)

type DeviceUsecase interface {
	CreateDevice(input *devicedto.CreateDeviceInput) error
	DeleteDevice(input *devicedto.DeleteDeviceInput) error
	EditDevice(input *devicedto.EditDeviceInput) error
	GetTraderDevices(input *devicedto.GetTraderDevicesInput) (*devicedto.GetTraderDevicesOutput, error)
	UpdateDeviceLiveness(deviceID string) error
	GetDeviceStatus(deviceID string) (*domain.Device, error)
	CheckOfflineDevices() error

	GetTraderDevicesStatus(traderID string) ([]*domain.Device, error)
}

type DefaultDeviceUsecase struct {
	deviceRepo domain.DeviceRepository
}

func NewDefaultDeviceUsecase(deviceRepo domain.DeviceRepository) *DefaultDeviceUsecase {
	return &DefaultDeviceUsecase{
		deviceRepo: deviceRepo,
	}
}

func (uc *DefaultDeviceUsecase) CreateDevice(input *devicedto.CreateDeviceInput) error {
	idGenerator, err := nanoid.Standard(15)
	if err != nil {
		return err
	}
	return uc.deviceRepo.CreateDevice(&domain.Device{
		DeviceID: idGenerator(),
		DeviceName: input.DeviceName,
		TraderID: input.TraderID,
		Enabled: input.Enabled,
	})
}

func (uc *DefaultDeviceUsecase) DeleteDevice(input *devicedto.DeleteDeviceInput) error {
	return uc.deviceRepo.DeleteDevice(input.DeviceID)
}

func (uc *DefaultDeviceUsecase) EditDevice(input *devicedto.EditDeviceInput) error {
	params := domain.UpdateDeviceParams{
		Name: input.DeviceName,
		Enabled: input.Enabled,
	}
	return uc.deviceRepo.UpdateDevice(input.DeviceID, params)
}

func (uc *DefaultDeviceUsecase) GetTraderDevices(input *devicedto.GetTraderDevicesInput) (*devicedto.GetTraderDevicesOutput, error) {
	devices, err := uc.deviceRepo.GetTraderDevices(input.TraderID)
	if err != nil {
		return nil, err
	}

	devicesOutput := make([]*devicedto.Device, len(devices))
	for i, device := range devices {
		devicesOutput[i] = &devicedto.Device{
			DeviceID: device.DeviceID,
			DeviceName: device.DeviceName,
			TraderID: device.TraderID,
			Enabled: device.Enabled,
		}
	}

	return &devicedto.GetTraderDevicesOutput{
		Devices: devicesOutput,
	}, nil
}

const DEVICE_OFFLINE_TIMEOUT = 2 * time.Minute

func (uc *DefaultDeviceUsecase) UpdateDeviceLiveness(deviceID string) error {
    now := time.Now()
    
    return uc.deviceRepo.UpdateDeviceLiveness(deviceID, now)
}

func (uc *DefaultDeviceUsecase) GetDeviceStatus(deviceID string) (*domain.Device, error) {
    return uc.deviceRepo.GetDeviceByID(deviceID)
}

// Background job для проверки оффлайн устройств
func (uc *DefaultDeviceUsecase) CheckOfflineDevices() error {
    threshold := time.Now().Add(-DEVICE_OFFLINE_TIMEOUT)
    
    return uc.deviceRepo.MarkDevicesOffline(threshold)
}

// GetTraderDevicesStatus получает статусы всех устройств трейдера
func (uc *DefaultDeviceUsecase) GetTraderDevicesStatus(traderID string) ([]*domain.Device, error) {
    return uc.deviceRepo.GetTraderDevices(traderID)
}