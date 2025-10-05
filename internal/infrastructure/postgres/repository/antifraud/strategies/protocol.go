package strategies

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/rules"
)

// ============= ИНТЕРФЕЙС СТРАТЕГИИ =============

// AntiFraudStrategy определяет интерфейс для стратегий проверки
type AntiFraudStrategy interface {
    Name() string
    Check(ctx context.Context, traderID string, rule *rules.AntiFraudRule) (*CheckResult, error)
    GetDescription() string
}

// CheckResult содержит результат проверки правила
type CheckResult struct {
    RuleName    string      `json:"rule_name"`
    Passed      bool        `json:"passed"`
    CurrentValue interface{} `json:"current_value"`
    Threshold   interface{} `json:"threshold"`
    Message     string      `json:"message"`
    Details     map[string]interface{} `json:"details,omitempty"`
}