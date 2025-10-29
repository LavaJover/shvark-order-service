package engine

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/rules"
    "gorm.io/gorm"
)

// RuleManager управляет правилами антифрода
type RuleManager struct {
    db *gorm.DB
}

func NewRuleManager(db *gorm.DB) *RuleManager {
    return &RuleManager{db: db}
}

// CreateRule создает новое правило
func (rm *RuleManager) CreateRule(ctx context.Context, name, ruleType string, config rules.RuleConfig, priority int) (*rules.AntiFraudRule, error) {
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    configMap := make(map[string]interface{})
    configBytes, _ := json.Marshal(config)
    json.Unmarshal(configBytes, &configMap)

    rule := &rules.AntiFraudRule{
        ID:       GenerateUUID(),
        Name:     name,
        Type:     ruleType,
        Config:   rules.JSONB(configMap), // Конвертируем в JSONB
        IsActive: true,
        Priority: priority,
    }

    if err := rm.db.WithContext(ctx).Create(rule).Error; err != nil {
        return nil, fmt.Errorf("failed to create rule: %w", err)
    }

    return rule, nil
}

// UpdateRule обновляет существующее правило
func (rm *RuleManager) UpdateRule(ctx context.Context, ruleID string, config rules.RuleConfig, isActive *bool, priority *int) error {
    updates := make(map[string]interface{})

    if config != nil {
        if err := config.Validate(); err != nil {
            return fmt.Errorf("invalid config: %w", err)
        }

        configMap := make(map[string]interface{})
        configBytes, _ := json.Marshal(config)
        json.Unmarshal(configBytes, &configMap)
        updates["config"] = rules.JSONB(configMap) // Конвертируем в JSONB
    }

    if isActive != nil {
        updates["is_active"] = *isActive
    }

    if priority != nil {
        updates["priority"] = *priority
    }

    updates["updated_at"] = time.Now()

    return rm.db.WithContext(ctx).
        Model(&rules.AntiFraudRule{}).
        Where("id = ?", ruleID).
        Updates(updates).Error
}

// GetRules получает все правила с фильтрацией
func (rm *RuleManager) GetRules(ctx context.Context, activeOnly bool) ([]rules.AntiFraudRule, error) {
    var rulesSlice []rules.AntiFraudRule
    query := rm.db.WithContext(ctx)

    if activeOnly {
        query = query.Where("is_active = ?", true)
    }

    err := query.Order("priority DESC").Find(&rulesSlice).Error
    return rulesSlice, err
}