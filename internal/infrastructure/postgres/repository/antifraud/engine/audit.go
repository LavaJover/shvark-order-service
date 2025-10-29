package engine

import (
    "database/sql/driver"
    "encoding/json"
    "errors"
    "time"

    "github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/strategies"
)

// CheckResultsJSON тип для хранения массива CheckResult в JSONB
type CheckResultsJSON []*strategies.CheckResult

// Value реализует интерфейс driver.Valuer
func (c CheckResultsJSON) Value() (driver.Value, error) {
    if c == nil {
        return nil, nil
    }
    return json.Marshal(c)
}

// Scan реализует интерфейс sql.Scanner
func (c *CheckResultsJSON) Scan(value interface{}) error {
    if value == nil {
        *c = nil
        return nil
    }

    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("type assertion to []byte failed")
    }

    var result []*strategies.CheckResult
    if err := json.Unmarshal(bytes, &result); err != nil {
        return err
    }

    *c = result
    return nil
}

// AntiFraudAuditLog для хранения истории проверок
type AntiFraudAuditLog struct {
    ID        string           `gorm:"primaryKey;type:uuid"`
    TraderID  string           `gorm:"not null;index"`
    CheckedAt time.Time        `gorm:"not null"`
    AllPassed bool             `gorm:"not null"`
    Results   CheckResultsJSON `gorm:"type:jsonb"`
    CreatedAt time.Time        `gorm:"default:CURRENT_TIMESTAMP"`
}