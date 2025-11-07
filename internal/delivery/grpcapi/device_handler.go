package grpcapi

import (
	"context"
	"log"

	"github.com/LavaJover/shvark-order-service/internal/usecase"
	devicedto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/device"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	if r.Params == nil {
		return nil, status.Error(codes.InvalidArgument, "params must be provided")
	}

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

func (h *DeviceHandler) UpdateDeviceLiveness(ctx context.Context, r *orderpb.UpdateDeviceLivenessRequest) (*orderpb.UpdateDeviceLivenessResponse, error) {
    if r.DeviceId == "" {
        return nil, status.Error(codes.InvalidArgument, "device_id is required")
    }
    
    err := h.deviceUc.UpdateDeviceLiveness(r.DeviceId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to update liveness: %v", err)
    }
    
    return &orderpb.UpdateDeviceLivenessResponse{
        Success: true,
    }, nil
}

func (h *DeviceHandler) GetDeviceStatus(ctx context.Context, r *orderpb.GetDeviceStatusRequest) (*orderpb.GetDeviceStatusResponse, error) {
    if r.DeviceId == "" {
        return nil, status.Error(codes.InvalidArgument, "device_id is required")
    }
    
    device, err := h.deviceUc.GetDeviceStatus(r.DeviceId)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "device not found: %v", err)
    }
    
    var lastPing int64
    if device.LastPingAt != nil {
        lastPing = device.LastPingAt.Unix()
    }
    
    return &orderpb.GetDeviceStatusResponse{
        DeviceId:  device.DeviceID,
        Online:    device.DeviceOnline,
        LastPing:  lastPing,
        Enabled:   device.Enabled,
    }, nil
}

// GetTraderDevicesStatus получает статусы всех устройств трейдера
func (h *DeviceHandler) GetTraderDevicesStatus(ctx context.Context, req *orderpb.GetTraderDevicesStatusRequest) (*orderpb.GetTraderDevicesStatusResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }
    
    log.Printf("📱 [GRPC-DEVICE] Getting devices status for trader: %s", req.TraderId)
    
    // Получаем устройства трейдера
    devices, err := h.deviceUc.GetTraderDevicesStatus(req.TraderId)
    if err != nil {
        log.Printf("❌ [GRPC-DEVICE] Failed to get trader devices: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to get trader devices: %v", err)
    }
    
    // Преобразуем в gRPC response
    deviceStatuses := make([]*orderpb.DeviceStatus, len(devices))
    for i, device := range devices {
        var lastPing int64
        if device.LastPingAt != nil {
            lastPing = device.LastPingAt.Unix()
        }
        
        deviceStatuses[i] = &orderpb.DeviceStatus{
            DeviceId:   device.DeviceID,
            DeviceName: device.DeviceName,
            Online:     device.DeviceOnline,
            LastPing:   lastPing,
            Enabled:    device.Enabled,
        }
    }
    
    log.Printf("✅ [GRPC-DEVICE] Retrieved %d devices for trader %s", len(devices), req.TraderId)
    
    return &orderpb.GetTraderDevicesStatusResponse{
        Devices: deviceStatuses,
    }, nil
}