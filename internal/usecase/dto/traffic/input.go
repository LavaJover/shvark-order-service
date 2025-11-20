package trafficdto

import "time"

type EditTrafficInput struct {
	ID 						string 					
	MerchantID 				*string 				
	TraderID 				*string 				
	TraderReward 			*float64 				
	TraderPriority 			*float64 				
	PlatformFee 			*float64				
	Enabled 				*bool
	Name					*string 					
	ActivityParams 			*TrafficActivityParams 	
	AntifraudParams 		*TrafficAntifraudParams 
	BusinessParams 			*TrafficBusinessParams 	
	// НОВОЕ: Конфигурация курсов
    ExchangeConfig  *ExchangeConfigInput `json:"exchange_config,omitempty"`
}

type TrafficActivityParams struct {
	MerchantUnlocked 	bool 
	TraderUnlocked   	bool 
	AntifraudUnlocked 	bool 
	ManuallyUnlocked  	bool 
}

type TrafficAntifraudParams struct {
	AntifraudRequired bool 
}

type TrafficBusinessParams struct {
	MerchantDealsDuration time.Duration
}

// Добавляем структуры для конфигурации курсов
type ExchangeConfigInput struct {
    ExchangeProvider   string    `json:"exchange_provider"`
    OrderBookRange     *OrderBookRangeInput `json:"order_book_range,omitempty"`
    MarkupPercent      float64   `json:"markup_percent"`
    FallbackProviders  []string  `json:"fallback_providers"`
    CurrencyPair       string    `json:"currency_pair"`
}

type OrderBookRangeInput struct {
    Start int `json:"start"`
    End   int `json:"end"`
}