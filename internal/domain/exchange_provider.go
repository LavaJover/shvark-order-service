// internal/domain/exchange_provider.go
package domain

import "context"

type ExchangeRateProvider interface {
    GetRate(ctx context.Context, config *ExchangeConfig) (float64, error)
    GetName() string
    IsHealthy(ctx context.Context) bool
}