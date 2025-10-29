package domain

import (
	"context"
	"fmt"
	"time"
)

// ============= КОНФИГУРАЦИЯ ПРАВИЛ =============

// AntiFraudRule представляет настраиваемое правило антифрода
type AntiFraudRule struct {
    ID          string                 `gorm:"primaryKey;type:uuid"`
    Name        string                 `gorm:"not null;unique"`
    Type        string                 `gorm:"not null"` // "consecutive_orders", "canceled_orders", "balance_threshold"
    Config      map[string]interface{} `gorm:"type:jsonb;not null"` // Настройки правила
    IsActive    bool                   `gorm:"default:true"`
    Priority    int                    `gorm:"default:0"` // Приоритет выполнения
    CreatedAt   time.Time              `gorm:"default:CURRENT_TIMESTAMP"`
    UpdatedAt   time.Time              `gorm:"default:CURRENT_TIMESTAMP"`
}

// RuleConfig представляет общий интерфейс для конфигурации правил
type RuleConfig interface {
    Validate() error
    GetThreshold() interface{}
}

// ConsecutiveOrdersConfig - конфигурация для правила подряд идущих сделок
type ConsecutiveOrdersConfig struct {
    MaxConsecutiveOrders int           `json:"max_consecutive_orders"`
    TimeWindow          time.Duration `json:"time_window"` // за какой период считать
    StatesToCount       []string      `json:"states_to_count"` // какие статусы считать
}

func (c *ConsecutiveOrdersConfig) Validate() error {
    if c.MaxConsecutiveOrders <= 0 {
        return fmt.Errorf("max_consecutive_orders must be positive")
    }
    if c.TimeWindow <= 0 {
        return fmt.Errorf("time_window must be positive")
    }
    return nil
}

func (c *ConsecutiveOrdersConfig) GetThreshold() interface{} {
    return c.MaxConsecutiveOrders
}

// CanceledOrdersConfig - конфигурация для правила отмененных сделок
type CanceledOrdersConfig struct {
    MaxCanceledOrders int           `json:"max_canceled_orders"`
    TimeWindow        time.Duration `json:"time_window"`
    CanceledStatuses  []string      `json:"canceled_statuses"`
}

func (c *CanceledOrdersConfig) Validate() error {
    if c.MaxCanceledOrders <= 0 {
        return fmt.Errorf("max_canceled_orders must be positive")
    }
    if c.TimeWindow <= 0 {
        return fmt.Errorf("time_window must be positive")
    }
    return nil
}

func (c *CanceledOrdersConfig) GetThreshold() interface{} {
    return c.MaxCanceledOrders
}

// BalanceThresholdConfig - конфигурация для правила минимального баланса
type BalanceThresholdConfig struct {
    MinBalance float64 `json:"min_balance"`
    Currency   string  `json:"currency"`
}

func (c *BalanceThresholdConfig) Validate() error {
    if c.MinBalance < 0 {
        return fmt.Errorf("min_balance cannot be negative")
    }
    if c.Currency == "" {
        return fmt.Errorf("currency is required")
    }
    return nil
}

func (c *BalanceThresholdConfig) GetThreshold() interface{} {
    return c.MinBalance
}

// ============= Отчеты =============

// type AntiFraudReport struct {
//     TraderID    string         `json:"trader_id"`
//     CheckedAt   time.Time      `json:"checked_at"`
//     AllPassed   bool           `json:"all_passed"`
//     Results     []*CheckResult `json:"results"`
//     FailedRules []string       `json:"failed_rules,omitempty"`
// }

type CheckResult struct {
    RuleName string                 `json:"rule_name"`
    Passed   bool                   `json:"passed"`
    Message  string                 `json:"message"`
    Details  map[string]interface{} `json:"details,omitempty"`
}

// ============= Правила =============

type AntiFraudRuleResponse struct {
    ID        string                 `json:"id"`
    Name      string                 `json:"name"`
    Type      string                 `json:"type"`
    Config    map[string]interface{} `json:"config"`
    IsActive  bool                   `json:"is_active"`
    Priority  int                    `json:"priority"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
}

type CreateRuleRequest struct {
    Name     string                 `json:"name"`
    Type     string                 `json:"type"`
    Config   map[string]interface{} `json:"config"`
    Priority int                    `json:"priority"`
}

func (r *CreateRuleRequest) Validate() error {
    if r.Name == "" {
        return fmt.Errorf("name is required")
    }
    if r.Type == "" {
        return fmt.Errorf("type is required")
    }
    if r.Config == nil {
        return fmt.Errorf("config is required")
    }
    return nil
}

type UpdateRuleRequest struct {
    RuleID   string                 `json:"rule_id"`
    Config   map[string]interface{} `json:"config,omitempty"`
    IsActive *bool                  `json:"is_active,omitempty"`
    Priority *int                   `json:"priority,omitempty"`
}

// ============= Аудит =============

type AuditLogResponse struct {
    ID        string         `json:"id"`
    TraderID  string         `json:"trader_id"`
    CheckedAt time.Time      `json:"checked_at"`
    AllPassed bool           `json:"all_passed"`
    Results   []*CheckResult `json:"results"`
    CreatedAt time.Time      `json:"created_at"`
}

type GetAuditLogsRequest struct {
    TraderID   string     `json:"trader_id,omitempty"`
    FromDate   *time.Time `json:"from_date,omitempty"`
    ToDate     *time.Time `json:"to_date,omitempty"`
    OnlyFailed bool       `json:"only_failed"`
    Limit      int        `json:"limit"`
    Offset     int        `json:"offset"`
}

type AntiFraudRepository interface {
    // Правила
    CreateRule(ctx context.Context, rule *AntiFraudRule) error
    UpdateRule(ctx context.Context, ruleID string, updates map[string]interface{}) error
    GetRules(ctx context.Context, activeOnly bool) ([]*AntiFraudRule, error)
    GetRuleByID(ctx context.Context, ruleID string) (*AntiFraudRule, error)
    DeleteRule(ctx context.Context, ruleID string) error
    
    // Аудит логи
    CreateAuditLog(ctx context.Context, log *AuditLog) error
    GetAuditLogs(ctx context.Context, filter *AuditLogFilter) ([]*AuditLog, error)
    GetTraderAuditHistory(ctx context.Context, traderID string, limit int) ([]*AuditLog, error)

    // Аудит разблокировок - НОВОЕ
    CreateUnlockAuditLog(ctx context.Context, log *UnlockAuditLog) error
    GetUnlockHistory(ctx context.Context, traderID string, limit int) ([]*UnlockAuditLog, error) // НОВОЕ
}

type AuditLog struct {
    ID        string
    TraderID  string
    CheckedAt time.Time
    AllPassed bool
    Results   []*CheckResult  // Изменено с interface{} на конкретный тип
    CreatedAt time.Time
}

type AuditLogFilter struct {
    TraderID   string
    FromDate   *time.Time
    ToDate     *time.Time
    OnlyFailed bool
    Limit      int
    Offset     int
}

type ManualUnlockRequest struct {
    TraderID         string `json:"trader_id"`
    AdminID          string `json:"admin_id"`
    Reason           string `json:"reason"`
    GracePeriodHours int    `json:"grace_period_hours"` // Длительность грейс-периода в часах
}

type AntiFraudReport struct {
    TraderID      string         `json:"trader_id"`
    CheckedAt     time.Time      `json:"checked_at"`
    AllPassed     bool           `json:"all_passed"`
    Results       []*CheckResult `json:"results"`
    FailedRules   []string       `json:"failed_rules,omitempty"`
    InGracePeriod bool           `json:"in_grace_period"`
}

// UnlockAuditLog для аудита ручных разблокировок
type UnlockAuditLog struct {
    ID               string
    TraderID         string
    AdminID          string
    Reason           string
    GracePeriodHours int
    UnlockedAt       time.Time
    CreatedAt        time.Time
}

// UnlockAuditLogResponse для API
type UnlockAuditLogResponse struct {
    ID               string    `json:"id"`
    TraderID         string    `json:"trader_id"`
    AdminID          string    `json:"admin_id"`
    Reason           string    `json:"reason"`
    GracePeriodHours int       `json:"grace_period_hours"`
    UnlockedAt       time.Time `json:"unlocked_at"`
    CreatedAt        time.Time `json:"created_at"`
}
