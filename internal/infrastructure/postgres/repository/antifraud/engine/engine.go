package engine

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
    "github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/rules"
    "github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/strategies"
    "gorm.io/gorm"
)

// AntiFraudEngine главный класс для работы с антифрод системой
type AntiFraudEngine struct {
    db              *gorm.DB
    strategies      map[string]strategies.AntiFraudStrategy
    logger          *slog.Logger
    snapshotManager *SnapshotManager // Добавили поле
}

func NewAntiFraudEngine(db *gorm.DB, logger *slog.Logger) *AntiFraudEngine {
    engine := &AntiFraudEngine{
        db:              db,
        strategies:      make(map[string]strategies.AntiFraudStrategy),
        logger:          logger,
        snapshotManager: NewSnapshotManager(db), // Инициализируем
    }

    return engine
}

// RegisterStrategy регистрирует новую стратегию
func (e *AntiFraudEngine) RegisterStrategy(strategy strategies.AntiFraudStrategy) {
    e.strategies[strategy.Name()] = strategy
    e.logger.Info("Registered antifraud strategy", "name", strategy.Name())
}

// CheckTrader проверяет трейдера по всем активным правилам с учётом грейс-периода
func (e *AntiFraudEngine) CheckTrader(ctx context.Context, traderID string) (*AntiFraudReport, error) {
    // Проверяем грейс-период
    inGracePeriod, err := e.snapshotManager.IsInGracePeriod(ctx, traderID)
    if err != nil {
        e.logger.Error("Failed to check grace period", "error", err)
    }

    if inGracePeriod {
        e.logger.Info("Trader is in grace period, skipping antifraud checks", "trader_id", traderID)
        return &AntiFraudReport{
            TraderID:      traderID,
            CheckedAt:     time.Now(),
            AllPassed:     true,
            Results:       []*strategies.CheckResult{},
            FailedRules:   []string{},
            InGracePeriod: true, // Добавили поле
        }, nil
    }

    // Получаем все активные правила
    var rulesList []rules.AntiFraudRule
    err = e.db.WithContext(ctx).
        Where("is_active = ?", true).
        Order("priority DESC").
        Find(&rulesList).Error

    if err != nil {
        return nil, fmt.Errorf("failed to fetch antifraud rules: %w", err)
    }

    report := &AntiFraudReport{
        TraderID:      traderID,
        CheckedAt:     time.Now(),
        Results:       make([]*strategies.CheckResult, 0, len(rulesList)),
        AllPassed:     true,
        InGracePeriod: false, // Добавили поле
    }

    // Проверяем каждое правило
    for _, rule := range rulesList {
        strategy, exists := e.strategies[rule.Type]
        if !exists {
            e.logger.Warn("Strategy not found for rule", "rule_type", rule.Type, "rule_name", rule.Name)
            continue
        }

        result, err := strategy.Check(ctx, traderID, &rule)
        if err != nil {
            e.logger.Error("Failed to check rule", "rule_name", rule.Name, "error", err)
            continue
        }

        report.Results = append(report.Results, result)

        if !result.Passed {
            report.AllPassed = false
            report.FailedRules = append(report.FailedRules, rule.Name)
        }
    }

    return report, nil
}

// AntiFraudReport содержит результат проверки трейдера
type AntiFraudReport struct {
    TraderID      string                    `json:"trader_id"`
    CheckedAt     time.Time                 `json:"checked_at"`
    AllPassed     bool                      `json:"all_passed"`
    Results       []*strategies.CheckResult `json:"results"`
    FailedRules   []string                  `json:"failed_rules,omitempty"`
    InGracePeriod bool                      `json:"in_grace_period"` // Добавили поле
}

// ProcessTraderCheck проверяет трейдера и обновляет статус трафика
func (e *AntiFraudEngine) ProcessTraderCheck(ctx context.Context, traderID string) error {
    report, err := e.CheckTrader(ctx, traderID)
    if err != nil {
        return fmt.Errorf("failed to check trader: %w", err)
    }

    // Если проверки не прошли, блокируем трафик
    if !report.AllPassed {
        err = e.updateTrafficStatus(ctx, traderID, false,
            fmt.Sprintf("Antifraud check failed: %v", report.FailedRules))
        if err != nil {
            return fmt.Errorf("failed to update traffic status: %w", err)
        }

        e.logger.Warn("Trader blocked by antifraud",
            "trader_id", traderID,
            "failed_rules", report.FailedRules)
    }

    // Сохраняем отчет для аудита
    if err := e.saveAuditLog(ctx, report); err != nil {
        e.logger.Error("Failed to save audit log", "error", err)
    }

    return nil
}

// updateTrafficStatus обновляет статус AntifraudUnlocked в TrafficModel
func (e *AntiFraudEngine) updateTrafficStatus(ctx context.Context, traderID string, unlocked bool, reason string) error {
    updates := map[string]interface{}{
        "antifraud_unlocked": unlocked,
        "updated_at":         time.Now(),
    }

    return e.db.WithContext(ctx).
        Model(&models.TrafficModel{}).
        Where("trader_id = ?", traderID).
        Updates(updates).Error
}

// saveAuditLog сохраняет результат проверки для аудита
func (e *AntiFraudEngine) saveAuditLog(ctx context.Context, report *AntiFraudReport) error {
    auditLog := &AntiFraudAuditLog{
        ID:        GenerateUUID(),
        TraderID:  report.TraderID,
        CheckedAt: report.CheckedAt,
        AllPassed: report.AllPassed,
        Results:   CheckResultsJSON(report.Results),
        CreatedAt: time.Now(),
    }

    return e.db.WithContext(ctx).Create(auditLog).Error
}
