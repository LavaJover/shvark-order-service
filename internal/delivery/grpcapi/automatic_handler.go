// internal/interfaces/grpcapi/automatic.go

package grpcapi

import (
    "context"
    "log"

    "github.com/LavaJover/shvark-order-service/internal/domain"
    "github.com/LavaJover/shvark-order-service/internal/usecase"
    orderpb "github.com/LavaJover/shvark-order-service/proto/gen/order"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type AutomaticHandler struct {
    automaticUc usecase.AutomaticUsecase
    orderpb.UnimplementedOrderServiceServer
}

func NewAutomaticHandler(automaticUc usecase.AutomaticUsecase) *AutomaticHandler {
    return &AutomaticHandler{
        automaticUc: automaticUc,
    }
}

// GetAutomaticLogs реализует gRPC метод для получения логов автоматики
// GetAutomaticLogs реализует gRPC метод для получения логов автоматики
func (h *AutomaticHandler) GetAutomaticLogs(ctx context.Context, req *orderpb.GetAutomaticLogsRequest) (*orderpb.GetAutomaticLogsResponse, error) {
    if req.Filter == nil {
        return nil, status.Error(codes.InvalidArgument, "filter is required")
    }
    
    // Преобразуем gRPC фильтр в доменный
    filter := &domain.AutomaticLogFilter{
        DeviceID:  req.Filter.DeviceId,
        TraderID:  req.Filter.TraderId,
        Action:    req.Filter.Action,
        Limit:     int(req.Filter.Limit),
        Offset:    int(req.Filter.Offset),
    }
    
    // Исправление: для optional bool используем прямое присвоение указателя
    if req.Filter.Success != nil {
        filter.Success = req.Filter.Success
    }
    
    if req.Filter.StartDate != nil {
        filter.StartDate = req.Filter.StartDate.AsTime()
    }
    
    if req.Filter.EndDate != nil {
        filter.EndDate = req.Filter.EndDate.AsTime()
    }
    
    logs, total, err := h.automaticUc.GetAutomaticLogs(ctx, filter)
    if err != nil {
        log.Printf("❌ [GRPC] Failed to get automatic logs: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to get automatic logs: %v", err)
    }
    
    // Преобразуем доменные логи в gRPC
    grpcLogs := make([]*orderpb.AutomaticLog, len(logs))
    for i, log := range logs {
        grpcLogs[i] = &orderpb.AutomaticLog{
            Id:             log.ID,
            DeviceId:       log.DeviceID,
            TraderId:       log.TraderID,
            OrderId:        log.OrderID,
            Amount:         log.Amount,
            PaymentSystem:  log.PaymentSystem,
            Direction:      log.Direction,
            Methods:        log.Methods,
            ReceivedAt:     timestamppb.New(log.ReceivedAt),
            Text:           log.Text,
            Action:         log.Action,
            Success:        log.Success,
            OrdersFound:    int32(log.OrdersFound),
            ErrorMessage:   log.ErrorMessage,
            ProcessingTime: log.ProcessingTime,
            BankName:       log.BankName,
            CardNumber:     log.CardNumber,
            CreatedAt:      timestamppb.New(log.CreatedAt),
        }
    }
    
    log.Printf("✅ [GRPC] Retrieved %d automatic logs (total: %d)", len(logs), total)
    
    return &orderpb.GetAutomaticLogsResponse{
        Logs:  grpcLogs,
        Total: int32(total),
    }, nil
}

// GetAutomaticStats получает статистику по автоматике
func (h *AutomaticHandler) GetAutomaticStats(ctx context.Context, req *orderpb.GetAutomaticStatsRequest) (*orderpb.GetAutomaticStatsResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }
    
    days := 7
    if req.Days > 0 {
        days = int(req.Days)
    }
    
    stats, err := h.automaticUc.GetAutomaticStats(ctx, req.TraderId, days)
    if err != nil {
        log.Printf("❌ [GRPC] Failed to get automatic stats: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to get automatic stats: %v", err)
    }
    
    // Преобразуем статистику устройств
    deviceStats := make(map[string]*orderpb.DeviceStats)
    for deviceID, ds := range stats.DeviceStats {
        deviceStats[deviceID] = &orderpb.DeviceStats{
            TotalAttempts: ds.TotalAttempts,
            SuccessCount:  ds.SuccessCount,
            SuccessRate:   ds.SuccessRate,
        }
    }
    
    log.Printf("✅ [GRPC] Retrieved automatic stats for trader %s", req.TraderId)
    
    return &orderpb.GetAutomaticStatsResponse{
        Stats: &orderpb.AutomaticStats{
            TotalAttempts:      stats.TotalAttempts,
            SuccessfulAttempts: stats.SuccessfulAttempts,
            ApprovedOrders:     stats.ApprovedOrders,
            NotFoundCount:      stats.NotFoundCount,
            FailedCount:        stats.FailedCount,
            AvgProcessingTime:  stats.AvgProcessingTime,
            DeviceStats:        deviceStats,
        },
    }, nil
}