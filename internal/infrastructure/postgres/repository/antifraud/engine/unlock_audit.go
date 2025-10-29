package engine

import "time"

// UnlockAuditLog для хранения истории ручных разблокировок
type UnlockAuditLog struct {
    ID               string    `gorm:"primaryKey;type:uuid"`
    TraderID         string    `gorm:"not null;index"`
    AdminID          string    `gorm:"not null"`
    Reason           string    `gorm:"type:text;not null"`
    GracePeriodHours int       `gorm:"not null"`
    UnlockedAt       time.Time `gorm:"not null"`
    CreatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (UnlockAuditLog) TableName() string {
    return "unlock_audit_logs"
}