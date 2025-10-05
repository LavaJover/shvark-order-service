package strategies

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/rules"
	"gorm.io/gorm"
)

// BalanceThresholdStrategy проверяет минимальный баланс трейдера
type BalanceThresholdStrategy struct {
    db             *gorm.DB
    balanceService BalanceService // Внешний сервис для получения баланса
}

// BalanceService интерфейс для получения баланса трейдера
type BalanceService interface {
    GetBalance(ctx context.Context, traderID string, currency string) (float64, error)
}

func NewBalanceThresholdStrategy(db *gorm.DB, balanceService BalanceService) *BalanceThresholdStrategy {
    return &BalanceThresholdStrategy{
        db:             db,
        balanceService: balanceService,
    }
}

func (s *BalanceThresholdStrategy) Name() string {
    return "balance_threshold"
}

func (s *BalanceThresholdStrategy) GetDescription() string {
    return "Проверка минимального баланса трейдера"
}

func (s *BalanceThresholdStrategy) Check(ctx context.Context, traderID string, rule *rules.AntiFraudRule) (*CheckResult, error) {
    configBytes, _ := json.Marshal(rule.Config)
    var config rules.BalanceThresholdConfig
    if err := json.Unmarshal(configBytes, &config); err != nil {
        return nil, fmt.Errorf("invalid config for balance_threshold rule: %w", err)
    }

    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    // Получаем текущий баланс трейдера
    currentBalance, err := s.balanceService.GetBalance(ctx, traderID, config.Currency)
    if err != nil {
        return nil, fmt.Errorf("failed to get trader balance: %w", err)
    }

    passed := currentBalance >= config.MinBalance

    return &CheckResult{
        RuleName:     rule.Name,
        Passed:       passed,
        CurrentValue: currentBalance,
        Threshold:    config.MinBalance,
        Message: fmt.Sprintf("Trader balance %.2f %s (minimum required: %.2f)", 
            currentBalance, config.Currency, config.MinBalance),
        Details: map[string]interface{}{
            "currency": config.Currency,
        },
    }, nil
}