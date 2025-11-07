// internal/usecase/automatic_usecase.go

package usecase

import (
    "context"

    "github.com/LavaJover/shvark-order-service/internal/domain"
)

type AutomaticUsecase interface {
    GetAutomaticLogs(ctx context.Context, filter *domain.AutomaticLogFilter) ([]*domain.AutomaticLog, int64, error)
    GetAutomaticStats(ctx context.Context, traderID string, days int) (*domain.AutomaticStats, error)
    GetRecentAutomaticActivity(ctx context.Context, traderID string, limit int) ([]*domain.AutomaticLog, error)
}

type DefaultAutomaticUsecase struct {
    orderRepo domain.OrderRepository
}

func NewDefaultAutomaticUsecase(orderRepo domain.OrderRepository) *DefaultAutomaticUsecase {
    return &DefaultAutomaticUsecase{
        orderRepo: orderRepo,
    }
}

func (uc *DefaultAutomaticUsecase) GetAutomaticLogs(ctx context.Context, filter *domain.AutomaticLogFilter) ([]*domain.AutomaticLog, int64, error) {
    logs, err := uc.orderRepo.GetAutomaticLogs(ctx, filter)
    if err != nil {
        return nil, 0, err
    }
    
    total, err := uc.orderRepo.GetAutomaticLogsCount(ctx, filter)
    if err != nil {
        return nil, 0, err
    }
    
    return logs, total, nil
}

func (uc *DefaultAutomaticUsecase) GetAutomaticStats(ctx context.Context, traderID string, days int) (*domain.AutomaticStats, error) {
    if days <= 0 {
        days = 7 // по умолчанию за неделю
    }
    
    stats, err := uc.orderRepo.GetAutomaticStats(ctx, traderID, days)
    if err != nil {
        return nil, err
    }
    
    // Вычисляем проценты успеха для каждого устройства
    for deviceID, deviceStats := range stats.DeviceStats {
        deviceStats.CalculateSuccessRate()
        stats.DeviceStats[deviceID] = deviceStats
    }
    
    return stats, nil
}

func (uc *DefaultAutomaticUsecase) GetRecentAutomaticActivity(ctx context.Context, traderID string, limit int) ([]*domain.AutomaticLog, error) {
    filter := &domain.AutomaticLogFilter{
        TraderID: traderID,
        Limit:    limit,
        Offset:   0,
    }
    
    logs, _, err := uc.GetAutomaticLogs(ctx, filter)
    return logs, err
}