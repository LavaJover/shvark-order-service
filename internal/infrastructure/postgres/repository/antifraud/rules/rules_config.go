package rules

import (
    "database/sql/driver"
    "encoding/json"
    "errors"
    "fmt"
    "time"
)

// JSONB тип для работы с JSONB полями PostgreSQL
type JSONB map[string]interface{}

// Value реализует интерфейс driver.Valuer
func (j JSONB) Value() (driver.Value, error) {
    if j == nil {
        return nil, nil
    }
    return json.Marshal(j)
}

// Scan реализует интерфейс sql.Scanner
func (j *JSONB) Scan(value interface{}) error {
    if value == nil {
        *j = nil
        return nil
    }

    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("type assertion to []byte failed")
    }

    result := make(map[string]interface{})
    if err := json.Unmarshal(bytes, &result); err != nil {
        return err
    }

    *j = result
    return nil
}

// AntiFraudRule представляет настраиваемое правило антифрода
type AntiFraudRule struct {
    ID        string    `gorm:"primaryKey;type:uuid"`
    Name      string    `gorm:"not null;unique"`
    Type      string    `gorm:"not null"`
    Config    JSONB     `gorm:"type:jsonb;not null"` // Используем custom JSONB тип
    IsActive  bool      `gorm:"default:true"`
    Priority  int       `gorm:"default:0"`
    CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
    UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// RuleConfig представляет общий интерфейс для конфигурации правил
type RuleConfig interface {
    Validate() error
    GetThreshold() interface{}
}

// ConsecutiveOrdersConfig - конфигурация для правила подряд идущих сделок
type ConsecutiveOrdersConfig struct {
    MaxConsecutiveOrders int           `json:"max_consecutive_orders"`
    TimeWindow           time.Duration `json:"time_window"`
    StatesToCount        []string      `json:"states_to_count"`
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