package engine

import (
    "context"
    "encoding/json"
    "time"

    "github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
    "gorm.io/gorm"
)

// UnlockSnapshot содержит состояние трейдера на момент разблокировки
type UnlockSnapshot struct {
    UnlockedAt          time.Time              `json:"unlocked_at"`
    UnlockedBy          string                 `json:"unlocked_by"` // Admin ID
    Reason              string                 `json:"reason"`
    FailedRules         []string               `json:"failed_rules"`
    Metrics             map[string]interface{} `json:"metrics"` // Текущие метрики
    GracePeriodDuration time.Duration          `json:"grace_period_duration"`
}

// SnapshotManager управляет снепшотами разблокировок
type SnapshotManager struct {
    db *gorm.DB
}

func NewSnapshotManager(db *gorm.DB) *SnapshotManager {
    return &SnapshotManager{db: db}
}

// CreateUnlockSnapshot создает снепшот при разблокировке
func (sm *SnapshotManager) CreateUnlockSnapshot(
    ctx context.Context,
    traderID string,
    adminID string,
    reason string,
    failedRules []string,
    metrics map[string]interface{},
    gracePeriodHours int,
) error {
    snapshot := UnlockSnapshot{
        UnlockedAt:          time.Now(),
        UnlockedBy:          adminID,
        Reason:              reason,
        FailedRules:         failedRules,
        Metrics:             metrics,
        GracePeriodDuration: time.Duration(gracePeriodHours) * time.Hour,
    }

    gracePeriodUntil := time.Now().Add(snapshot.GracePeriodDuration)

    snapshotJSON, _ := json.Marshal(snapshot)
    var snapshotMap map[string]interface{}
    json.Unmarshal(snapshotJSON, &snapshotMap)

    return sm.db.WithContext(ctx).
        Model(&models.TrafficModel{}).
        Where("trader_id = ?", traderID).
        Updates(map[string]interface{}{
            "antifraud_unlocked": true,
            "updated_at":         time.Now(),
            "grace_period_until": gracePeriodUntil,
        }).Error
}

// IsInGracePeriod проверяет, действует ли грейс-период
func (sm *SnapshotManager) IsInGracePeriod(ctx context.Context, traderID string) (bool, error) {
    var count int64
    err := sm.db.WithContext(ctx).
        Model(&models.TrafficModel{}).
        Where("trader_id = ? AND grace_period_until > ?", traderID, time.Now()).
        Count(&count).Error
    
    if err != nil {
        return false, err
    }

    return count > 0, nil
}

// ResetGracePeriod сбрасывает грейс-период
func (sm *SnapshotManager) ResetGracePeriod(ctx context.Context, traderID string) error {
    return sm.db.WithContext(ctx).
        Model(&models.TrafficModel{}).
        Where("trader_id = ?", traderID).
        Updates(map[string]interface{}{
            "grace_period_until": nil,
            "updated_at":         time.Now(),
        }).Error
}