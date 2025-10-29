package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/engine"
	"github.com/google/uuid"
)

type AntiFraudUseCase interface {
    // Проверка трейдера
    CheckTrader(ctx context.Context, traderID string) (*domain.AntiFraudReport, error)
    ProcessTraderCheck(ctx context.Context, traderID string) error
    
    // Управление правилами
    CreateRule(ctx context.Context, req *domain.CreateRuleRequest) (*domain.AntiFraudRuleResponse, error)
    UpdateRule(ctx context.Context, req *domain.UpdateRuleRequest) error
    GetRules(ctx context.Context, activeOnly bool) ([]*domain.AntiFraudRuleResponse, error)
    GetRule(ctx context.Context, ruleID string) (*domain.AntiFraudRuleResponse, error)
    DeleteRule(ctx context.Context, ruleID string) error
    
    // Аудит
    GetAuditLogs(ctx context.Context, req *domain.GetAuditLogsRequest) ([]*domain.AuditLogResponse, error)
    GetTraderAuditHistory(ctx context.Context, traderID string, limit int) ([]*domain.AuditLogResponse, error)

	// Добавляем новые методы
	ManualUnlock(ctx context.Context, req *domain.ManualUnlockRequest) error
	ResetGracePeriod(ctx context.Context, traderID string) error

	GetUnlockHistory(ctx context.Context, traderID string, limit int) ([]*domain.UnlockAuditLogResponse, error) // НОВОЕ
}

type antiFraudUseCase struct {
    engine *engine.AntiFraudEngine
    repo   domain.AntiFraudRepository
	snapshotManager *engine.SnapshotManager // Добавили поле
}

func NewAntiFraudUseCase(
    engine *engine.AntiFraudEngine,
    repo domain.AntiFraudRepository,
	snapshotManager *engine.SnapshotManager, // Добавили параметр
) AntiFraudUseCase {
    return &antiFraudUseCase{
        engine: engine,
        repo:   repo,
    }
}

// ============= Проверка трейдеров =============

func (uc *antiFraudUseCase) CheckTrader(ctx context.Context, traderID string) (*domain.AntiFraudReport, error) {
    if traderID == "" {
        return nil, fmt.Errorf("trader_id is required")
    }

    report, err := uc.engine.CheckTrader(ctx, traderID)
    if err != nil {
        return nil, fmt.Errorf("failed to check trader: %w", err)
    }

    return uc.convertEngineToDomainReport(report), nil
}

func (uc *antiFraudUseCase) ProcessTraderCheck(ctx context.Context, traderID string) error {
    if traderID == "" {
        return fmt.Errorf("trader_id is required")
    }

    return uc.engine.ProcessTraderCheck(ctx, traderID)
}

// ============= Управление правилами =============

func (uc *antiFraudUseCase) CreateRule(ctx context.Context, req *domain.CreateRuleRequest) (*domain.AntiFraudRuleResponse, error) {
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("validation error: %w", err)
    }

    rule := &domain.AntiFraudRule{
        ID:       uuid.New().String(),
        Name:     req.Name,
        Type:     req.Type,
        Config:   req.Config,
        IsActive: true,
        Priority: req.Priority,
    }

    if err := uc.repo.CreateRule(ctx, rule); err != nil {
        return nil, fmt.Errorf("failed to create rule: %w", err)
    }

    return uc.convertRuleToDomainResponse(rule), nil
}

func (uc *antiFraudUseCase) UpdateRule(ctx context.Context, req *domain.UpdateRuleRequest) error {
    if req.RuleID == "" {
        return fmt.Errorf("rule_id is required")
    }

    updates := make(map[string]interface{})

    if req.Config != nil {
        updates["config"] = req.Config
    }

    if req.IsActive != nil {
        updates["is_active"] = *req.IsActive
    }

    if req.Priority != nil {
        updates["priority"] = *req.Priority
    }

    return uc.repo.UpdateRule(ctx, req.RuleID, updates)
}

func (uc *antiFraudUseCase) GetRules(ctx context.Context, activeOnly bool) ([]*domain.AntiFraudRuleResponse, error) {
    rules, err := uc.repo.GetRules(ctx, activeOnly)
    if err != nil {
        return nil, fmt.Errorf("failed to get rules: %w", err)
    }

    result := make([]*domain.AntiFraudRuleResponse, 0, len(rules))
    for _, rule := range rules {
        result = append(result, uc.convertRuleToDomainResponse(rule))
    }

    return result, nil
}

func (uc *antiFraudUseCase) GetRule(ctx context.Context, ruleID string) (*domain.AntiFraudRuleResponse, error) {
    if ruleID == "" {
        return nil, fmt.Errorf("rule_id is required")
    }

    rule, err := uc.repo.GetRuleByID(ctx, ruleID)
    if err != nil {
        return nil, fmt.Errorf("failed to get rule: %w", err)
    }

    return uc.convertRuleToDomainResponse(rule), nil
}

func (uc *antiFraudUseCase) DeleteRule(ctx context.Context, ruleID string) error {
    if ruleID == "" {
        return fmt.Errorf("rule_id is required")
    }

    return uc.repo.DeleteRule(ctx, ruleID)
}

// ============= Аудит =============

func (uc *antiFraudUseCase) GetAuditLogs(ctx context.Context, req *domain.GetAuditLogsRequest) ([]*domain.AuditLogResponse, error) {
    filter := &domain.AuditLogFilter{
        TraderID:   req.TraderID,
        FromDate:   req.FromDate,
        ToDate:     req.ToDate,
        OnlyFailed: req.OnlyFailed,
        Limit:      req.Limit,
        Offset:     req.Offset,
    }

    if filter.Limit <= 0 {
        filter.Limit = 50
    }

    logs, err := uc.repo.GetAuditLogs(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("failed to get audit logs: %w", err)
    }

    result := make([]*domain.AuditLogResponse, 0, len(logs))
    for _, log := range logs {
        result = append(result, uc.convertAuditLogToResponse(log))
    }

    return result, nil
}

func (uc *antiFraudUseCase) GetTraderAuditHistory(ctx context.Context, traderID string, limit int) ([]*domain.AuditLogResponse, error) {
    if traderID == "" {
        return nil, fmt.Errorf("trader_id is required")
    }

    if limit <= 0 {
        limit = 10
    }

    logs, err := uc.repo.GetTraderAuditHistory(ctx, traderID, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to get trader audit history: %w", err)
    }

    result := make([]*domain.AuditLogResponse, 0, len(logs))
    for _, log := range logs {
        result = append(result, uc.convertAuditLogToResponse(log))
    }

    return result, nil
}

// ============= Вспомогательные функции =============

func (uc *antiFraudUseCase) convertEngineToDomainReport(engineReport *engine.AntiFraudReport) *domain.AntiFraudReport {
    results := make([]*domain.CheckResult, 0, len(engineReport.Results))
    for _, r := range engineReport.Results {
        results = append(results, &domain.CheckResult{
            RuleName: r.RuleName,
            Passed:   r.Passed,
            Message:  r.Message,
            Details:  r.Details,
        })
    }

    return &domain.AntiFraudReport{
        TraderID:    engineReport.TraderID,
        CheckedAt:   engineReport.CheckedAt,
        AllPassed:   engineReport.AllPassed,
        Results:     results,
        FailedRules: engineReport.FailedRules,
    }
}

func (uc *antiFraudUseCase) convertRuleToDomainResponse(rule *domain.AntiFraudRule) *domain.AntiFraudRuleResponse {
    return &domain.AntiFraudRuleResponse{
        ID:        rule.ID,
        Name:      rule.Name,
        Type:      rule.Type,
        Config:    rule.Config,
        IsActive:  rule.IsActive,
        Priority:  rule.Priority,
        CreatedAt: rule.CreatedAt,
        UpdatedAt: rule.UpdatedAt,
    }
}

func (uc *antiFraudUseCase) convertAuditLogToResponse(log *domain.AuditLog) *domain.AuditLogResponse {
    // Теперь не нужна проверка типа, так как Results уже правильного типа
    return &domain.AuditLogResponse{
        ID:        log.ID,
        TraderID:  log.TraderID,
        CheckedAt: log.CheckedAt,
        AllPassed: log.AllPassed,
        Results:   log.Results,
        CreatedAt: log.CreatedAt,
    }
}

// internal/usecase/antifraud_usecase.go

// ManualUnlock вручную разблокирует трейдера с грейс-периодом
func (uc *antiFraudUseCase) ManualUnlock(ctx context.Context, req *domain.ManualUnlockRequest) error {
    if req.TraderID == "" {
        return fmt.Errorf("trader_id is required")
    }

    if req.AdminID == "" {
        return fmt.Errorf("admin_id is required")
    }

    if req.GracePeriodHours <= 0 {
        req.GracePeriodHours = 24
    }

    // Получаем текущий отчет о проверке
    report, err := uc.engine.CheckTrader(ctx, req.TraderID)
    if err != nil {
        return fmt.Errorf("failed to check trader: %w", err)
    }

    // Собираем текущие метрики для снепшота
    metrics := make(map[string]interface{})
    for _, result := range report.Results {
        metrics[result.RuleName] = result.Details
    }

    // Создаём снепшот разблокировки
    err = uc.snapshotManager.CreateUnlockSnapshot(
        ctx,
        req.TraderID,
        req.AdminID,
        req.Reason,
        report.FailedRules,
        metrics,
        req.GracePeriodHours,
    )

    if err != nil {
        return fmt.Errorf("failed to create unlock snapshot: %w", err)
    }

    // НОВОЕ: Сохраняем в аудит-лог разблокировку
    unlockLog := &domain.UnlockAuditLog{
        ID:               uuid.New().String(),
        TraderID:         req.TraderID,
        AdminID:          req.AdminID,
        Reason:           req.Reason,
        GracePeriodHours: req.GracePeriodHours,
        UnlockedAt:       time.Now(),
    }

    if err := uc.repo.CreateUnlockAuditLog(ctx, unlockLog); err != nil {
        // Не критично если не удалось сохранить в аудит
        // Основная разблокировка уже произошла
        return fmt.Errorf("trader unlocked but failed to save audit log: %w", err)
    }

    return nil
}

// ResetGracePeriod сбрасывает грейс-период
func (uc *antiFraudUseCase) ResetGracePeriod(ctx context.Context, traderID string) error {
    if uc.snapshotManager == nil {
        return fmt.Errorf("CRITICAL: snapshotManager is nil in usecase")
    }
    return uc.snapshotManager.ResetGracePeriod(ctx, traderID)
}

// GetUnlockHistory получает историю ручных разблокировок
func (uc *antiFraudUseCase) GetUnlockHistory(ctx context.Context, traderID string, limit int) ([]*domain.UnlockAuditLogResponse, error) {
    if traderID == "" {
        return nil, fmt.Errorf("trader_id is required")
    }

    if limit <= 0 {
        limit = 20
    }

    logs, err := uc.repo.GetUnlockHistory(ctx, traderID, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to get unlock history: %w", err)
    }

    result := make([]*domain.UnlockAuditLogResponse, 0, len(logs))
    for _, log := range logs {
        result = append(result, &domain.UnlockAuditLogResponse{
            ID:               log.ID,
            TraderID:         log.TraderID,
            AdminID:          log.AdminID,
            Reason:           log.Reason,
            GracePeriodHours: log.GracePeriodHours,
            UnlockedAt:       log.UnlockedAt,
            CreatedAt:        log.CreatedAt,
        })
    }

    return result, nil
}