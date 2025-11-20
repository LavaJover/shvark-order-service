package background

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/infrastructure/usdt"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
)

type BackgroundTasks struct {
    OrderUsecase    usecase.OrderUsecase
    DisputeUsecase  usecase.DisputeUsecase
    DeviceUsecase   usecase.DeviceUsecase
    TrafficUsecase      usecase.TrafficUsecase
    ExchangeRateService usecase.ExchangeRateService
}

func NewBackgroundTasks(
    orderUC usecase.OrderUsecase, 
    disputeUC usecase.DisputeUsecase, 
    deviceUC usecase.DeviceUsecase,
    trafficUC usecase.TrafficUsecase,
    exchangeService usecase.ExchangeRateService,
) *BackgroundTasks {
    return &BackgroundTasks{
        OrderUsecase:        orderUC,
        DisputeUsecase:      disputeUC,
        DeviceUsecase:       deviceUC,
        TrafficUsecase:      trafficUC,
        ExchangeRateService: exchangeService,
    }
}

func (bt *BackgroundTasks) StartAll(ctx context.Context) {
    go bt.startOrderAutoCancel(ctx)
    go bt.startExchangeRatesSystem(ctx) // ЗАМЕНЯЕМ старый метод на новый
    go bt.startAutoAcceptExpiredDisputes(ctx)
    go bt.startDeviceOfflineCheck(ctx)
    go bt.startExchangeHealthMonitoring(ctx) // НОВАЯ задача мониторинга
    go bt.startExchangeRatesWarmup(ctx)      // НОВАЯ задача предзагрузки
}

func (bt *BackgroundTasks) startOrderAutoCancel(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := bt.OrderUsecase.CancelExpiredOrders(ctx); err != nil {
                log.Printf("Auto-cancel error: %v\n", err)
            }
        }
    }
}

func (bt *BackgroundTasks) startCryptoRatesUpdate(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            usdtRate, err := usdt.GET_USDT_RUB_RATES(5)
            if err != nil {
                log.Printf("USD/RUB rates update failed: %v", err)
                continue
            }
            log.Printf("USD/RUB rates updated: usdt/rub=%.2f", usdtRate)
        }
    }
}

func (bt *BackgroundTasks) startAutoAcceptExpiredDisputes(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := bt.DisputeUsecase.AcceptExpiredDisputes(); err != nil {
                log.Printf("Auto-accept dispute error: %v\n", err)
            }
        }
    }
}

func (bt *BackgroundTasks) startDeviceOfflineCheck(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := bt.DeviceUsecase.CheckOfflineDevices(); err != nil {
                log.Printf("Error checking offline devices: %v", err)
            }
        }
    }
}

// НОВЫЙ МЕТОД: Комплексная система курсов
func (bt *BackgroundTasks) startExchangeRatesSystem(ctx context.Context) {
    // Основное обновление курсов - каждые 2 минуты
    mainTicker := time.NewTicker(2 * time.Minute)
    // Быстрое обновление при ошибках - каждые 30 секунд
    fastTicker := time.NewTicker(30 * time.Second)
    
    defer mainTicker.Stop()
    defer fastTicker.Stop()
    
    var lastError bool
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-mainTicker.C:
            if err := bt.updateAllActiveExchangeRates(ctx); err != nil {
                slog.Error("Failed to update exchange rates", "error", err)
                lastError = true
            } else {
                lastError = false
            }
        case <-fastTicker.C:
            // Быстрое обновление только если были ошибки
            if lastError {
                if err := bt.updateAllActiveExchangeRates(ctx); err != nil {
                    slog.Error("Failed to update exchange rates (fast retry)", "error", err)
                } else {
                    lastError = false
                    slog.Info("Exchange rates recovered after fast retry")
                }
            }
        }
    }
}

// Обновляет курсы для всех активных трафиков
func (bt *BackgroundTasks) updateAllActiveExchangeRates(ctx context.Context) error {
    // Получаем первую страницу активных трафиков
    traffics, err := bt.TrafficUsecase.GetTrafficRecords(1, 100)
    if err != nil {
        return err
    }

    var successCount, errorCount int
    
    for _, traffic := range traffics {
        if !traffic.Enabled {
            continue
        }

        config, err := traffic.GetExchangeConfig()
        if err != nil {
            slog.Warn("Failed to get exchange config for traffic", 
                "trafficID", traffic.ID, 
                "error", err)
            errorCount++
            continue
        }

        // Получаем курс (это обновит кеш)
        _, provider, err := bt.ExchangeRateService.GetRateWithFallback(ctx, config)
        if err != nil {
            slog.Warn("Failed to update exchange rate for traffic", 
                "trafficID", traffic.ID, 
                "provider", config.ExchangeProvider,
                "error", err)
            errorCount++
        } else {
            successCount++
            slog.Debug("Exchange rate updated", 
                "trafficID", traffic.ID, 
                "provider", provider,
                "currencyPair", config.CurrencyPair)
        }
    }

    slog.Info("Exchange rates update completed", 
        "successful", successCount, 
        "errors", errorCount,
        "total", len(traffics))
    
    return nil
}

// НОВАЯ ЗАДАЧА: Мониторинг здоровья бирж
func (bt *BackgroundTasks) startExchangeHealthMonitoring(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            bt.checkExchangeHealth(ctx)
        }
    }
}

func (bt *BackgroundTasks) checkExchangeHealth(ctx context.Context) {
    errors := bt.ExchangeRateService.HealthCheck(ctx)
    
    // Логируем проблемы
    for provider, err := range errors {
        slog.Error("Exchange provider unhealthy", 
            "provider", provider, 
            "error", err)
        
        // Можно отправить алерт в систему мониторинга
        bt.sendHealthAlert(provider, err)
    }

    // Логируем успешные проверки
    if len(errors) == 0 {
        slog.Debug("All exchange providers are healthy")
    }

    // Отчет метрик
    healthyCount := len(bt.ExchangeRateService.GetAvailableProviders()) - len(errors)
    slog.Info("Exchange health status", 
        "healthy_providers", healthyCount,
        "unhealthy_providers", len(errors))
}

// НОВАЯ ЗАДАЧА: Предварительная загрузка курсов для кеширования
func (bt *BackgroundTasks) startExchangeRatesWarmup(ctx context.Context) {
    // Запускаем сразу при старте
    bt.warmupExchangeRates(ctx)
    
    // Затем каждые 5 минут
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            bt.warmupExchangeRates(ctx)
        }
    }
}

func (bt *BackgroundTasks) warmupExchangeRates(ctx context.Context) {
    traffics, err := bt.TrafficUsecase.GetTrafficRecords(1, 50) // Берем первые 50
    if err != nil {
        slog.Error("Failed to get traffics for warmup", "error", err)
        return
    }

    var warmedUp int
    for _, traffic := range traffics {
        if !traffic.Enabled {
            continue
        }

        config, err := traffic.GetExchangeConfig()
        if err != nil {
            slog.Warn("Failed to get exchange config for warmup", 
                "trafficID", traffic.ID, 
                "error", err)
            continue
        }

        // Пробуем получить курс (для кеширования)
        _, _, err = bt.ExchangeRateService.GetRateWithFallback(ctx, config)
        if err != nil {
            slog.Warn("Failed to warmup exchange rate", 
                "trafficID", traffic.ID, 
                "provider", config.ExchangeProvider,
                "error", err)
        } else {
            warmedUp++
        }
    }

    slog.Info("Exchange rates warmup completed", 
        "warmed_up", warmedUp, 
        "total_processed", len(traffics))
}

func (bt *BackgroundTasks) sendHealthAlert(provider string, err error) {
    // Реализация отправки алерта (например, в Telegram, Slack, или систему мониторинга)
    // Это может быть отдельный сервис нотификаций
    slog.Warn("EXCHANGE HEALTH ALERT", 
        "provider", provider, 
        "error", err.Error())
}