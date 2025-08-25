package grpcapi

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/usecase"
	devicedto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/device"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
)

type DeviceHandler struct {
	deviceUc usecase.DeviceUsecase
	orderpb.UnimplementedDeviceServiceServer
}

func NewDeviceHandler(deviceUc usecase.DeviceUsecase) *DeviceHandler {
	return &DeviceHandler{
		deviceUc: deviceUc,
	}
}

func (h *DeviceHandler) CreateDevice(ctx context.Context, r *orderpb.CreateDeviceRequest) (*orderpb.CreateDeviceResponse, error) {
	createDeviceInput := devicedto.CreateDeviceInput{
		DeviceName: r.DeviceName,
		TraderID: r.TraderId,
		Enabled: r.Enabled,
	}

	err := h.deviceUc.CreateDevice(&createDeviceInput)
	if err != nil {
		return nil, err
	}

	return &orderpb.CreateDeviceResponse{}, nil
}

func (h *DeviceHandler) GetTraderDevices(ctx context.Context, r *orderpb.GetTraderDevicesRequest) (*orderpb.GetTraderDevicesResponse, error) {
	getTraderDevicesInput := devicedto.GetTraderDevicesInput{
		TraderID: r.TraderId,
	}
	output, err := h.deviceUc.GetTraderDevices(&getTraderDevicesInput)
	if err != nil {
		return nil, err
	}

	devices := make([]*orderpb.Device, len(output.Devices))
	for i, device := range output.Devices {
		devices[i] = &orderpb.Device{
			DeviceId: device.DeviceID,
			DeviceName: device.DeviceName,
			TraderId: device.TraderID,
			Enabled: device.Enabled,
		}
	}

	return &orderpb.GetTraderDevicesResponse{
		Devices: devices,
	}, nil
}

func (h *DeviceHandler) DeleteDevice(ctx context.Context, r *orderpb.DeleteDeviceRequest) (*orderpb.DeleteDeviceResponse, error) {
	deleteDeviceInput := devicedto.DeleteDeviceInput{
		DeviceID: r.DeviceId,
	}
	err := h.deviceUc.DeleteDevice(&deleteDeviceInput)
	if err != nil {
		return nil, err
	}

	return &orderpb.DeleteDeviceResponse{}, nil
}

func (h *DeviceHandler) EditDevice(ctx context.Context, r *orderpb.EditDeviceRequest) (*orderpb.EditDeviceResponse, error) {
	editDeviceInput := devicedto.EditDeviceInput{
		DeviceID: r.DeviceId,
		DeviceName: r.Params.DeviceName,
		Enabled: r.Params.Enabled,
	}
	err := h.deviceUc.EditDevice(&editDeviceInput)
	if err != nil {
		return nil, err
	}

	return &orderpb.EditDeviceResponse{}, nil
}