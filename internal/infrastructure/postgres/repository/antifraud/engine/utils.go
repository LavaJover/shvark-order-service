package engine

import (
	"context"
	"log/slog"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/rules"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/strategies"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============= ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ =============

func GenerateUUID() string {
    // Здесь должна быть реальная генерация UUID
    return uuid.New().String()
}

// Пример использования системы
func ExampleUsage() {
    // Инициализация базы данных (предполагается, что db уже настроена)
    var db *gorm.DB
    logger := slog.Default()

    // Создаем движок антифрода
    engine := NewAntiFraudEngine(db, logger)

    // Регистрируем стратегии
    engine.RegisterStrategy(strategies.NewConsecutiveOrdersStrategy(db))
    engine.RegisterStrategy(strategies.NewCanceledOrdersStrategy(db))

    // Для стратегии баланса нужен сервис баланса
    // balanceService := NewBalanceService() // реализация зависит от вашей архитектуры
    // engine.RegisterStrategy(NewBalanceThresholdStrategy(db, balanceService))

    // Создаем менеджер правил
    ruleManager := NewRuleManager(db)

    // Создаем правила
    consecutiveConfig := &rules.ConsecutiveOrdersConfig{
        MaxConsecutiveOrders: 10,
        TimeWindow:          24 * time.Hour,
        StatesToCount:       []string{"COMPLETED", "PROCESSING"},
    }

    ruleManager.CreateRule(context.Background(), 
        "Max Consecutive Orders", 
        "consecutive_orders", 
        consecutiveConfig, 
        100)

    canceledConfig := &rules.CanceledOrdersConfig{
        MaxCanceledOrders: 5,
        TimeWindow:        24 * time.Hour,
        CanceledStatuses:  []string{"CANCELED", "REJECTED"},
    }

    ruleManager.CreateRule(context.Background(), 
        "Max Canceled Orders", 
        "canceled_orders", 
        canceledConfig, 
        90)

    // Запускаем планировщик для автоматических проверок
    scheduler := NewScheduler(engine, db, 30*time.Minute, logger)
    go scheduler.Start(context.Background())

    // Проверка конкретного трейдера
    report, err := engine.CheckTrader(context.Background(), "trader-123")
    if err != nil {
        logger.Error("Failed to check trader", "error", err)
        return
    }

    if !report.AllPassed {
        logger.Warn("Trader failed antifraud checks", 
            "trader_id", report.TraderID,
            "failed_rules", report.FailedRules)
    }
}