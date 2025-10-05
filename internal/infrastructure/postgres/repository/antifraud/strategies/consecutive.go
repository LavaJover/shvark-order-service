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

// ============= КОНКРЕТНЫЕ СТРАТЕГИИ =============

// ConsecutiveOrdersStrategy проверяет количество сделок подряд
type ConsecutiveOrdersStrategy struct {
    db *gorm.DB
}

func NewConsecutiveOrdersStrategy(db *gorm.DB) *ConsecutiveOrdersStrategy {
    return &ConsecutiveOrdersStrategy{db: db}
}

func (s *ConsecutiveOrdersStrategy) Name() string {
    return "consecutive_orders"
}

func (s *ConsecutiveOrdersStrategy) GetDescription() string {
    return "Проверка максимального количества последовательных сделок"
}

func (s *ConsecutiveOrdersStrategy) Check(ctx context.Context, traderID string, rule *rules.AntiFraudRule) (*CheckResult, error) {
    // Парсим конфигурацию
    configBytes, _ := json.Marshal(rule.Config)
    var config rules.ConsecutiveOrdersConfig
    if err := json.Unmarshal(configBytes, &config); err != nil {
        return nil, fmt.Errorf("invalid config for consecutive_orders rule: %w", err)
    }

    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    // Считаем количество последовательных сделок
    var consecutiveCount int64
    timeLimit := time.Now().Add(-config.TimeWindow)

    query := s.db.WithContext(ctx).Model(&models.OrderModel{}).
        Where("trader_id = ?", traderID).
        Where("created_at >= ?", timeLimit)

    if len(config.StatesToCount) > 0 {
        query = query.Where("status IN ?", config.StatesToCount)
    }

    if err := query.Count(&consecutiveCount).Error; err != nil {
        return nil, fmt.Errorf("failed to count consecutive orders: %w", err)
    }

    passed := consecutiveCount <= int64(config.MaxConsecutiveOrders)

    return &CheckResult{
        RuleName:     rule.Name,
        Passed:       passed,
        CurrentValue: consecutiveCount,
        Threshold:    config.MaxConsecutiveOrders,
        Message: fmt.Sprintf("Trader has %d consecutive orders in last %v (limit: %d)", 
            consecutiveCount, config.TimeWindow, config.MaxConsecutiveOrders),
        Details: map[string]interface{}{
            "time_window": config.TimeWindow.String(),
            "states_counted": config.StatesToCount,
        },
    }, nil
}