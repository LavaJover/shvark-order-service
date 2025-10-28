package repository

import (
    "context"
    "time"

    "github.com/LavaJover/shvark-order-service/internal/domain"
    "github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/engine"
    "github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/rules"
    "github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/strategies"
    "gorm.io/gorm"
)

type antiFraudRepository struct {
    db *gorm.DB
}

func NewAntiFraudRepository(db *gorm.DB) domain.AntiFraudRepository {
    return &antiFraudRepository{db: db}
}

// ============= Правила =============

func (r *antiFraudRepository) CreateRule(ctx context.Context, rule *domain.AntiFraudRule) error {
    dbRule := &rules.AntiFraudRule{
        ID:       rule.ID,
        Name:     rule.Name,
        Type:     rule.Type,
        Config:   rule.Config,
        IsActive: rule.IsActive,
        Priority: rule.Priority,
    }

    return r.db.WithContext(ctx).Create(dbRule).Error
}

func (r *antiFraudRepository) UpdateRule(ctx context.Context, ruleID string, updates map[string]interface{}) error {
    updates["updated_at"] = time.Now()

    return r.db.WithContext(ctx).
        Model(&rules.AntiFraudRule{}).
        Where("id = ?", ruleID).
        Updates(updates).Error
}

func (r *antiFraudRepository) GetRules(ctx context.Context, activeOnly bool) ([]*domain.AntiFraudRule, error) {
    var dbRules []rules.AntiFraudRule
    query := r.db.WithContext(ctx)

    if activeOnly {
        query = query.Where("is_active = ?", true)
    }

    err := query.Order("priority DESC").Find(&dbRules).Error
    if err != nil {
        return nil, err
    }

    result := make([]*domain.AntiFraudRule, 0, len(dbRules))
    for _, dbRule := range dbRules {
        result = append(result, r.convertDBRuleToDomain(&dbRule))
    }

    return result, nil
}

func (r *antiFraudRepository) GetRuleByID(ctx context.Context, ruleID string) (*domain.AntiFraudRule, error) {
    var dbRule rules.AntiFraudRule
    err := r.db.WithContext(ctx).Where("id = ?", ruleID).First(&dbRule).Error
    if err != nil {
        return nil, err
    }

    return r.convertDBRuleToDomain(&dbRule), nil
}

func (r *antiFraudRepository) DeleteRule(ctx context.Context, ruleID string) error {
    return r.db.WithContext(ctx).Delete(&rules.AntiFraudRule{}, "id = ?", ruleID).Error
}

// ============= Аудит логи =============

func (r *antiFraudRepository) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
    // Конвертируем domain.CheckResult в strategies.CheckResult
    strategyResults := make([]*strategies.CheckResult, len(log.Results))
    for i, res := range log.Results {
        strategyResults[i] = &strategies.CheckResult{
            RuleName: res.RuleName,
            Passed:   res.Passed,
            Message:  res.Message,
            Details:  res.Details,
        }
    }

    dbLog := &engine.AntiFraudAuditLog{
        ID:        log.ID,
        TraderID:  log.TraderID,
        CheckedAt: log.CheckedAt,
        AllPassed: log.AllPassed,
        Results:   strategyResults,
        CreatedAt: time.Now(),
    }

    return r.db.WithContext(ctx).Create(dbLog).Error
}

func (r *antiFraudRepository) GetAuditLogs(ctx context.Context, filter *domain.AuditLogFilter) ([]*domain.AuditLog, error) {
    query := r.db.WithContext(ctx).Model(&engine.AntiFraudAuditLog{})

    if filter.TraderID != "" {
        query = query.Where("trader_id = ?", filter.TraderID)
    }

    if filter.FromDate != nil {
        query = query.Where("checked_at >= ?", filter.FromDate)
    }

    if filter.ToDate != nil {
        query = query.Where("checked_at <= ?", filter.ToDate)
    }

    if filter.OnlyFailed {
        query = query.Where("all_passed = ?", false)
    }

    var dbLogs []engine.AntiFraudAuditLog
    err := query.
        Order("checked_at DESC").
        Limit(filter.Limit).
        Offset(filter.Offset).
        Find(&dbLogs).Error

    if err != nil {
        return nil, err
    }

    result := make([]*domain.AuditLog, 0, len(dbLogs))
    for _, dbLog := range dbLogs {
        domainLog, err := r.convertDBAuditLogToDomain(&dbLog)
        if err != nil {
            // Логируем ошибку, но продолжаем
            continue
        }
        result = append(result, domainLog)
    }

    return result, nil
}

func (r *antiFraudRepository) GetTraderAuditHistory(ctx context.Context, traderID string, limit int) ([]*domain.AuditLog, error) {
    if limit <= 0 {
        limit = 10
    }

    var dbLogs []engine.AntiFraudAuditLog
    err := r.db.WithContext(ctx).
        Where("trader_id = ?", traderID).
        Order("checked_at DESC").
        Limit(limit).
        Find(&dbLogs).Error

    if err != nil {
        return nil, err
    }

    result := make([]*domain.AuditLog, 0, len(dbLogs))
    for _, dbLog := range dbLogs {
        domainLog, err := r.convertDBAuditLogToDomain(&dbLog)
        if err != nil {
            // Логируем ошибку, но продолжаем
            continue
        }
        result = append(result, domainLog)
    }

    return result, nil
}

// ============= Вспомогательные функции =============

func (r *antiFraudRepository) convertDBRuleToDomain(dbRule *rules.AntiFraudRule) *domain.AntiFraudRule {
    return &domain.AntiFraudRule{
        ID:        dbRule.ID,
        Name:      dbRule.Name,
        Type:      dbRule.Type,
        Config:    dbRule.Config,
        IsActive:  dbRule.IsActive,
        Priority:  dbRule.Priority,
        CreatedAt: dbRule.CreatedAt,
        UpdatedAt: dbRule.UpdatedAt,
    }
}

func (r *antiFraudRepository) convertDBAuditLogToDomain(dbLog *engine.AntiFraudAuditLog) (*domain.AuditLog, error) {
    // Конвертируем strategies.CheckResult в domain.CheckResult
    domainResults := make([]*domain.CheckResult, 0, len(dbLog.Results))
    
    for _, res := range dbLog.Results {
        domainResults = append(domainResults, &domain.CheckResult{
            RuleName: res.RuleName,
            Passed:   res.Passed,
            Message:  res.Message,
            Details:  res.Details,
        })
    }

    return &domain.AuditLog{
        ID:        dbLog.ID,
        TraderID:  dbLog.TraderID,
        CheckedAt: dbLog.CheckedAt,
        AllPassed: dbLog.AllPassed,
        Results:   domainResults,
        CreatedAt: dbLog.CreatedAt,
    }, nil
}