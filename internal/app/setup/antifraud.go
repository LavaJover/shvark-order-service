package setup

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/engine"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/rules"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/strategies"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
)

type AntiFraudSystem struct {
    Engine      *engine.AntiFraudEngine
    Scheduler   *engine.Scheduler
    RuleManager *engine.RuleManager
    UseCase     usecase.AntiFraudUseCase
}

func InitializeAntiFraud(deps *Dependencies) (*AntiFraudSystem, error) {
    antifraudLogger := slog.Default()
    
    // Создаем engine точно как в исходном коде
    antifraudEngine := engine.NewAntiFraudEngine(deps.DB, antifraudLogger)
    
    // Получаем snapshotManager из engine (как в исходном коде)
    snapshotManager := antifraudEngine.GetSnapshotManager()
    if snapshotManager == nil {
        log.Fatal("CRITICAL: snapshotManager is nil after GetSnapshotManager()")
    }
    log.Printf("✓ SnapshotManager initialized successfully: %p", snapshotManager)

    // Регистрируем стратегии
    antifraudEngine.RegisterStrategy(strategies.NewConsecutiveOrdersStrategy(deps.DB))
    antifraudEngine.RegisterStrategy(strategies.NewCanceledOrdersStrategy(deps.DB))

    // Создаем repository и use case
    antiFraudRepo := repository.NewAntiFraudRepository(deps.DB)
    antiFraudUseCase := usecase.NewAntiFraudUseCase(antifraudEngine, antiFraudRepo, snapshotManager)

    // Создаем rule manager и настраиваем правила
    ruleManager := engine.NewRuleManager(deps.DB)
    if err := setupDefaultRules(context.Background(), ruleManager); err != nil {
        return nil, err
    }

    // Создаем scheduler
    scheduler := engine.NewScheduler(antifraudEngine, deps.DB, 1*time.Minute, antifraudLogger)

    return &AntiFraudSystem{
        Engine:      antifraudEngine,
        Scheduler:   scheduler,
        RuleManager: ruleManager,
        UseCase:     antiFraudUseCase,
    }, nil
}

func setupDefaultRules(ctx context.Context, ruleManager *engine.RuleManager) error {
    consecutiveConfig := &rules.ConsecutiveOrdersConfig{
        MaxConsecutiveOrders: 10,
        TimeWindow:          30 * time.Minute,
        StatesToCount:       []string{"CANCELED"},
    }

    if _, err := ruleManager.CreateRule(ctx, 
        "Max Consecutive Orders", 
        "consecutive_orders", 
        consecutiveConfig, 
        100); err != nil {
            slog.Error("failed to create consecutive_orders rule", "error", err.Error())
    }

    canceledConfig := &rules.CanceledOrdersConfig{
        MaxCanceledOrders: 5,
        TimeWindow:        30 * time.Minute,
        CanceledStatuses:  []string{"CANCELED"},
    }

    if _, err := ruleManager.CreateRule(ctx, 
        "Max Canceled Orders", 
        "canceled_orders", 
        canceledConfig, 
    	90); err != nil {
			slog.Error("failed to create canceled_orders rule", "error", err.Error())
		}
	return nil
}