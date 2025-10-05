package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"gorm.io/gorm"
)

// ============= ПЛАНИРОВЩИК ПРОВЕРОК =============

// Scheduler для автоматических проверок
type Scheduler struct {
    engine   *AntiFraudEngine
    db       *gorm.DB
    interval time.Duration
    logger   *slog.Logger
}

func NewScheduler(engine *AntiFraudEngine, db *gorm.DB, interval time.Duration, logger *slog.Logger) *Scheduler {
    return &Scheduler{
        engine:   engine,
        db:       db,
        interval: interval,
        logger:   logger,
    }
}

// Start запускает планировщик
func (s *Scheduler) Start(ctx context.Context) {
    ticker := time.NewTicker(s.interval)
    defer ticker.Stop()

    s.logger.Info("Starting antifraud scheduler", "interval", s.interval)

    for {
        select {
        case <-ctx.Done():
            s.logger.Info("Stopping antifraud scheduler")
            return
        case <-ticker.C:
            if err := s.runChecks(ctx); err != nil {
                s.logger.Error("Failed to run scheduled checks", "error", err)
            }
        }
    }
}

// runChecks выполняет проверки для всех активных трейдеров
func (s *Scheduler) runChecks(ctx context.Context) error {
    // Получаем всех активных трейдеров
    var traderIDs []string
    err := s.db.WithContext(ctx).
        Model(&models.TrafficModel{}).
        Where("antifraud_required = ?", true).
        Pluck("trader_id", &traderIDs).Error

    if err != nil {
        s.logger.Error("failed to get active traders")
        return fmt.Errorf("failed to get active traders: %w", err)
    }

    s.logger.Info("Running scheduled antifraud checks", "traders_count", len(traderIDs))

    // Проверяем каждого трейдера
    for _, traderID := range traderIDs {
        if err := s.engine.ProcessTraderCheck(ctx, traderID); err != nil {
            s.logger.Error("Failed to check trader", "trader_id", traderID, "error", err)
        }
    }

    return nil
}