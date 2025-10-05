package engine

import (
	"time"

	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/strategies"
)

// AntiFraudAuditLog для хранения истории проверок
type AntiFraudAuditLog struct {
    ID        string         `gorm:"primaryKey;type:uuid"`
    TraderID  string         `gorm:"not null;index"`
    CheckedAt time.Time      `gorm:"not null"`
    AllPassed bool           `gorm:"not null"`
    Results   []*strategies.CheckResult `gorm:"type:jsonb"`
    CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
}