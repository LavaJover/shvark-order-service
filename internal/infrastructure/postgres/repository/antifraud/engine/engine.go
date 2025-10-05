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

// ============= ОСНОВНОЙ ДВИЖОК АНТИФРОДА =============

// AntiFraudEngine главный класс для работы с антифрод системой
type AntiFraudEngine struct {
    db         *gorm.DB
    strategies map[string]strategies.AntiFraudStrategy
    logger     *slog.Logger
}

func NewAntiFraudEngine(db *gorm.DB, logger *slog.Logger) *AntiFraudEngine {
    engine := &AntiFraudEngine{
        db:         db,
        strategies: make(map[string]strategies.AntiFraudStrategy),
        logger:     logger,
    }

    return engine
}

// RegisterStrategy регистрирует новую стратегию
func (e *AntiFraudEngine) RegisterStrategy(strategy strategies.AntiFraudStrategy) {
    e.strategies[strategy.Name()] = strategy
    e.logger.Info("Registered antifraud strategy", "name", strategy.Name())
}

// CheckTrader проверяет трейдера по всем активным правилам
func (e *AntiFraudEngine) CheckTrader(ctx context.Context, traderID string) (*AntiFraudReport, error) {
    // Получаем все активные правила
    var rules []rules.AntiFraudRule
    err := e.db.WithContext(ctx).
        Where("is_active = ?", true).
        Order("priority DESC").
        Find(&rules).Error

    if err != nil {
        return nil, fmt.Errorf("failed to fetch antifraud rules: %w", err)
    }

    report := &AntiFraudReport{
        TraderID:    traderID,
        CheckedAt:   time.Now(),
        Results:     make([]*strategies.CheckResult, 0, len(rules)),
        AllPassed:   true,
    }

    // Проверяем каждое правило
    for _, rule := range rules {
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
    TraderID    string         `json:"trader_id"`
    CheckedAt   time.Time      `json:"checked_at"`
    AllPassed   bool           `json:"all_passed"`
    Results     []*strategies.CheckResult `json:"results"`
    FailedRules []string       `json:"failed_rules,omitempty"`
}

// ProcessTraderCheck проверяет трейдера и обновляет статус трафика
func (e *AntiFraudEngine) ProcessTraderCheck(ctx context.Context, traderID string) error {
    // Проверяем трейдера
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
        "reason":            reason,
        "updated_at":        time.Now(),
    }

    if unlocked {
        updates["unlocked_at"] = time.Now()
    } else {
        updates["locked_at"] = time.Now()
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
        Results:   report.Results,
        CreatedAt: time.Now(),
    }

    return e.db.WithContext(ctx).Create(auditLog).Error
}