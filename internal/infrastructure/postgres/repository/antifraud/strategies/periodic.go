package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/rules"
	"gorm.io/gorm"
)

// CanceledOrdersStrategy проверяет количество отмененных сделок
type CanceledOrdersStrategy struct {
    db *gorm.DB
}

func NewCanceledOrdersStrategy(db *gorm.DB) *CanceledOrdersStrategy {
    return &CanceledOrdersStrategy{db: db}
}

func (s *CanceledOrdersStrategy) Name() string {
    return "canceled_orders"
}

func (s *CanceledOrdersStrategy) GetDescription() string {
    return "Проверка количества отмененных сделок за период"
}

func (s *CanceledOrdersStrategy) Check(ctx context.Context, traderID string, rule *rules.AntiFraudRule) (*CheckResult, error) {
    configBytes, _ := json.Marshal(rule.Config)
    var config rules.CanceledOrdersConfig
    if err := json.Unmarshal(configBytes, &config); err != nil {
        return nil, fmt.Errorf("invalid config for canceled_orders rule: %w", err)
    }

    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    var canceledCount int64
    timeLimit := time.Now().Add(-config.TimeWindow)

    err := s.db.WithContext(ctx).Model(&models.OrderModel{}).
        Where("trader_id = ?", traderID).
        Where("status IN ?", config.CanceledStatuses).
        Where("updated_at >= ?", timeLimit).
        Count(&canceledCount).Error

    if err != nil {
        return nil, fmt.Errorf("failed to count canceled orders: %w", err)
    }

    passed := canceledCount <= int64(config.MaxCanceledOrders)

    return &CheckResult{
        RuleName:     rule.Name,
        Passed:       passed,
        CurrentValue: canceledCount,
        Threshold:    config.MaxCanceledOrders,
        Message: fmt.Sprintf("Trader has %d canceled orders in last %v (limit: %d)", 
            canceledCount, config.TimeWindow, config.MaxCanceledOrders),
        Details: map[string]interface{}{
            "time_window": config.TimeWindow.String(),
            "canceled_statuses": config.CanceledStatuses,
        },
    }, nil
}