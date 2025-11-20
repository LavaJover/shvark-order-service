// internal/infrastructure/exchange_providers/rapira_provider.go
package infrastructure

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "github.com/LavaJover/shvark-order-service/internal/domain"
)

type RapiraProvider struct {
    client *http.Client
}

type RapiraItem struct {
	Price 	float64 `json:"price"`
	Amount  float64 `json:"amount"`
}

type RapiraResponse struct {
	Ask struct {
		Direction 	 string  	`json:"direction"`
		Symbol 		 string  	`json:"symbol"`
		MaxAmount	 float64 	`json:"max_amount"`
		MinAmount	 float64 	`json:"min_amount"`
		HighestPrice float64 	`json:"highest_price"`
		LowestPrice  float64 	`json:"lowest_price"`
		Items 		 []RapiraItem `json:"items"`
	}
}

func NewRapiraProvider() *RapiraProvider {
    return &RapiraProvider{
        client: &http.Client{
            Timeout: 5 * time.Second,
        },
    }
}

func (r *RapiraProvider) GetName() string {
    return "rapira"
}

func (r *RapiraProvider) GetRate(ctx context.Context, config *domain.ExchangeConfig) (float64, error) {
    url := fmt.Sprintf("https://api.rapira.net/market/exchange-plate-mini?symbol=%s", config.CurrencyPair)
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return 0, fmt.Errorf("failed to create request: %w", err)
    }
    
    resp, err := r.client.Do(req)
    if err != nil {
        return 0, fmt.Errorf("failed to get rates from Rapira: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return 0, fmt.Errorf("rapira API returned status: %d", resp.StatusCode)
    }
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return 0, fmt.Errorf("failed to read response body: %w", err)
    }
    
    var rapiraResponse RapiraResponse
    if err := json.Unmarshal(body, &rapiraResponse); err != nil {
        return 0, fmt.Errorf("failed to parse Rapira response: %w", err)
    }
    
    // Применяем настройки позиций стакана
    positions := config.OrderBookPositions
    if positions == nil {
        positions = &domain.OrderBookRange{Start: 0, End: 4} // значения по умолчанию
    }
    
    return r.calculateAveragePrice(rapiraResponse.Ask.Items, positions.Start, positions.End)
}

func (r *RapiraProvider) calculateAveragePrice(items []RapiraItem, start, end int) (float64, error) {
    if len(items) == 0 {
        return 0, fmt.Errorf("no items in order book")
    }
    
    if start < 0 || end >= len(items) || start > end {
        return 0, fmt.Errorf("invalid positions range: start=%d, end=%d, available=%d", start, end, len(items))
    }
    
    total := 0.0
    count := 0
    
    for i := start; i <= end && i < len(items); i++ {
        total += items[i].Price
        count++
    }
    
    if count == 0 {
        return 0, fmt.Errorf("no valid positions in range")
    }
    
    return total / float64(count), nil
}

func (r *RapiraProvider) IsHealthy(ctx context.Context) bool {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    
    _, err := r.GetRate(ctx, &domain.ExchangeConfig{CurrencyPair: "USDT/RUB"})
    return err == nil
}