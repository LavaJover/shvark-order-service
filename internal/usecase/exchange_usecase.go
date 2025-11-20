// internal/usecase/exchange_service.go
package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	infrastructure "github.com/LavaJover/shvark-order-service/internal/infrastructure/exchange_providers"
)

type ExchangeRateService interface {
    GetRate(ctx context.Context, config *domain.ExchangeConfig) (float64, error)
    GetRateWithFallback(ctx context.Context, config *domain.ExchangeConfig) (float64, string, error)
    GetAvailableProviders() []string
    HealthCheck(ctx context.Context) map[string]error
}

type DefaultExchangeRateService struct {
    providers map[string]domain.ExchangeRateProvider
    cache     *ExchangeRateCache
}

type ExchangeRateCache struct {
    rates map[string]CachedRate
    ttl   time.Duration
    mu    sync.RWMutex
}

type CachedRate struct {
    rate      float64
    timestamp time.Time
    provider  string
}

func NewDefaultExchangeRateService() *DefaultExchangeRateService {
    service := &DefaultExchangeRateService{
        providers: make(map[string]domain.ExchangeRateProvider),
        cache: &ExchangeRateCache{
            rates: make(map[string]CachedRate),
            ttl:   10 * time.Second, // Кешируем на 10 секунд
        },
    }
    
    // Регистрируем провайдеры
    service.RegisterProvider("rapira", infrastructure.NewRapiraProvider())
    // service.RegisterProvider("bybit", infrastructure.NewByBitProvider())
    // service.RegisterProvider("binance", infrastructure.NewBinanceProvider())
    // service.RegisterProvider("manual", infrastructure.NewManualRateProvider())
    
    return service
}

func (s *DefaultExchangeRateService) RegisterProvider(name string, provider domain.ExchangeRateProvider) {
    s.providers[name] = provider
}

func (s *DefaultExchangeRateService) GetRate(ctx context.Context, config *domain.ExchangeConfig) (float64, error) {
    return s.getRateWithProvider(ctx, config, config.ExchangeProvider)
}

func (s *DefaultExchangeRateService) GetRateWithFallback(ctx context.Context, config *domain.ExchangeConfig) (float64, string, error) {
    // Пробуем основной провайдер
    rate, err := s.GetRate(ctx, config)
    if err == nil {
        return rate, config.ExchangeProvider, nil
    }
    
    // Пробуем fallback провайдеры
    for _, providerName := range config.FallbackProviders {
        if providerName == config.ExchangeProvider {
            continue // Пропускаем основной, он уже пробовался
        }
        
        rate, err := s.getRateWithProvider(ctx, config, providerName)
        if err == nil {
            // Логируем использование fallback
            slog.Warn("Using fallback exchange provider", 
                "primary", config.ExchangeProvider, 
                "fallback", providerName,
                "error", err)
            return rate, providerName, nil
        }
    }
    
    return 0, "", fmt.Errorf("all exchange providers failed: %w", err)
}

func (s *DefaultExchangeRateService) getRateWithProvider(ctx context.Context, config *domain.ExchangeConfig, providerName string) (float64, error) {
    // Проверяем кеш
    cacheKey := fmt.Sprintf("%s_%s", providerName, config.CurrencyPair)
    if cached, ok := s.cache.Get(cacheKey); ok {
        return cached.rate, nil
    }
    
    provider, exists := s.providers[providerName]
    if !exists {
        return 0, fmt.Errorf("exchange provider %s not found", providerName)
    }
    
    // Получаем курс от провайдера
    baseRate, err := provider.GetRate(ctx, config)
    if err != nil {
        return 0, err
    }
    
    // Применяем наценку
    finalRate := baseRate * (1 + config.MarkupPercent/100)
    
    // Сохраняем в кеш
    s.cache.Set(cacheKey, finalRate, providerName)
    
    return finalRate, nil
}

func (s *DefaultExchangeRateService) GetAvailableProviders() []string {
    providers := make([]string, 0, len(s.providers))
    for name := range s.providers {
        providers = append(providers, name)
    }
    return providers
}

func (s *DefaultExchangeRateService) HealthCheck(ctx context.Context) map[string]error {
    errors := make(map[string]error)
    
    for name, provider := range s.providers {
        if _, err := provider.GetRate(ctx, &domain.ExchangeConfig{CurrencyPair: "USDT/RUB"}); err != nil {
            errors[name] = err
        }
    }
    
    return errors
}

// Методы кеша
func (c *ExchangeRateCache) Get(key string) (CachedRate, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    cached, exists := c.rates[key]
    if !exists || time.Since(cached.timestamp) > c.ttl {
        return CachedRate{}, false
    }
    
    return cached, true
}

func (c *ExchangeRateCache) Set(key string, rate float64, provider string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.rates[key] = CachedRate{
        rate:      rate,
        timestamp: time.Now(),
        provider:  provider,
    }
}