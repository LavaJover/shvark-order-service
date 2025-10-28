package usecase

import (
    "context"
    "fmt"

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
}

type antiFraudUseCase struct {
    engine *engine.AntiFraudEngine
    repo   domain.AntiFraudRepository
}

func NewAntiFraudUseCase(
    engine *engine.AntiFraudEngine,
    repo domain.AntiFraudRepository,
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