package cascade

// import (
// 	"context"
// 	"fmt"
// 	"log/slog"
// 	"math/rand/v2"
// 	"sort"
// 	"sync"
// 	"time"

// 	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
// 	"github.com/LavaJover/shvark-order-service/internal/domain"
// 	metrics "github.com/rcrowley/go-metrics"
// 	"gorm.io/gorm"
// )

// // Каскадный engine для подбора реквизитов
// type CascadeEngine struct {
//     filters        []Filter
//     scorer         Scorer
//     cache          Cache
//     walletService  handlers.HTTPWalletHandler
//     trafficRepo    domain.TrafficRepository
// }

// // Фильтры применяются последовательно, от быстрых к медленным
// type Filter interface {
//     Filter(ctx context.Context, req *MatchRequest, candidates []*BankDetail) ([]*BankDetail, error)
//     Priority() int // Приоритет выполнения (чем меньше - тем раньше)
// }

// // Scorer для взвешенного выбора
// type Scorer interface {
//     Score(bankDetail *domain.BankDetail, traffic *domain.Traffic) float64
//     Select(candidates []*domain.BankDetail) *domain.BankDetail
// }

// type Cache struct {

// }

// // ----------------------------------------- Оптимизированные структуры данных

// // Денормализованная модель для быстрого поиска
// type BankDetailView struct {
//     ID                    string    `json:"id"`
//     TraderID              string    `json:"trader_id"`
//     Currency              string    `json:"currency"`
//     PaymentSystem         string    `json:"payment_system"`
//     MinAmount            float64   `json:"min_amount"`
//     MaxAmount            float64   `json:"max_amount"`
//     BankCode             string    `json:"bank_code"`
//     NspkCode             string    `json:"nspk_code"`
//     Enabled              bool      `json:"enabled"`
// 	MaxOrdersSimultaneosly  int32
// 	MaxAmountDay			float64
// 	MaxAmountMonth  		float64
// 	MaxQuantityDay			int32
// 	MaxQuantityMonth		int32
// 	MinOrderAmount 			float32
// 	MaxOrderAmount 			float32
// 	Delay 					time.Duration
    
//     // Денормализованные поля из Traffic
//     TrafficEnabled       bool      `json:"traffic_enabled"`
//     TraderPriority       float64   `json:"trader_priority"`
//     TraderRewardPercent  float64   `json:"trader_reward_percent"`
//     PlatformFee          float64   `json:"platform_fee"`
    
//     // Денормализованные счетчики
//     CurrentPendingCount  int32     `json:"current_pending_count"`
//     TodayOrdersCount     int32     `json:"today_orders_count"`
//     TodayOrdersAmount    float64   `json:"today_orders_amount"`
//     MonthOrdersCount     int32     `json:"month_orders_count"`
//     MonthOrdersAmount    float64   `json:"month_orders_amount"`
//     LastCompletedAt      *time.Time `json:"last_completed_at"`
    
//     // Кэшированный баланс (обновляется раз в 30 сек)
//     CachedBalance        float64   `json:"cached_balance"`
//     BalanceUpdatedAt     time.Time `json:"balance_updated_at"`
// }

// // In-memory кэш для быстрого доступа
// type BankDetailCache struct {
//     byCurrencyPS     map[string]map[string][]*BankDetailView // currency -> payment_system -> []bank_details
//     byTrader         map[string]*BankDetailView
//     mutex            sync.RWMutex
//     lastUpdated      time.Time
// }


// // ------------------------------- Каскадные фильтры

// // 1. Быстрый статический фильтр (in-memory)
// type StaticFilter struct {
//     cache *BankDetailCache
// }

// func (f *StaticFilter) Filter(ctx context.Context, req *MatchRequest, candidates []*BankDetailView) ([]*BankDetailView, error) {
//     // Используем предварительно отфильтрованные данные из кэша
//     key := fmt.Sprintf("%s:%s", req.Currency, req.PaymentSystem)
    
//     f.cache.mutex.RLock()
//     defer f.cache.mutex.RUnlock()
    
//     currencyMap, ok := f.cache.byCurrencyPS[req.Currency]
//     if !ok {
//         return nil, nil
//     }
    
//     candidates, ok = currencyMap[req.PaymentSystem]
//     if !ok {
//         return nil, nil
//     }
    
//     // Фильтрация по сумме и банковским кодам
//     var filtered []*BankDetailView
//     for _, bd := range candidates {
//         if bd.Enabled && bd.TrafficEnabled &&
//            req.AmountFiat >= bd.MinAmount && req.AmountFiat <= bd.MaxAmount {
            
//             if req.BankCode == "" || bd.BankCode == req.BankCode {
//                 if req.NspkCode == "" || bd.NspkCode == req.NspkCode {
//                     filtered = append(filtered, bd)
//                 }
//             }
//         }
//     }
    
//     return filtered, nil
// }

// func (f *StaticFilter) Priority() int { return 1 }

// // 2. Фильтр по лимитам (быстрая проверка денормализованных данных)
// type LimitsFilter struct{}

// func (f *LimitsFilter) Filter(ctx context.Context, req *MatchRequest, candidates []*BankDetailView) ([]*BankDetailView, error) {
//     var filtered []*BankDetailView
    
//     for _, bd := range candidates {
//         if bd.CurrentPendingCount >= bd.MaxOrdersSimultaneosly {
//             continue
//         }
//         if bd.TodayOrdersCount+1 > bd.MaxQuantityDay {
//             continue
//         }
//         if bd.TodayOrdersAmount+req.AmountFiat > bd.MaxAmountDay {
//             continue
//         }
//         if bd.MonthOrdersCount+1 > bd.MaxQuantityMonth {
//             continue
//         }
//         if bd.MonthOrdersAmount+req.AmountFiat > bd.MaxAmountMonth {
//             continue
//         }
        
//         // Проверка задержки
//         if bd.LastCompletedAt != nil {
//             requiredDelay := time.Duration(bd.Delay) * time.Second
//             if time.Since(*bd.LastCompletedAt) < requiredDelay {
//                 continue
//             }
//         }
        
//         filtered = append(filtered, bd)
//     }
    
//     return filtered, nil
// }

// func (f *LimitsFilter) Priority() int { return 2 }

// // 3. Фильтр по балансу (с кэшированием)
// type BalanceFilter struct {
//     walletService handlers.HTTPWalletHandler
// }

// func (f *BalanceFilter) Filter(ctx context.Context, req *MatchRequest, candidates []*BankDetailView) ([]*BankDetailView, error) {
//     if len(candidates) == 0 {
//         return candidates, nil
//     }
    
//     // Группируем по трейдерам для пакетного запроса
//     traderIDs := make([]string, 0, len(candidates))
//     traderMap := make(map[string][]*BankDetailView)
    
//     for _, bd := range candidates {
//         traderMap[bd.TraderID] = append(traderMap[bd.TraderID], bd)
//         traderIDs = append(traderIDs, bd.TraderID)
//     }
    
//     // Пакетный запрос балансов
//     balances, err := f.walletService.GetBatchBalances(ctx, traderIDs)
//     if err != nil {
//         return nil, err
//     }
    
//     var filtered []*BankDetailView
//     for traderID, traderCandidates := range traderMap {
//         balance, exists := balances[traderID]
//         if !exists || balance < req.AmountCrypto {
//             continue
//         }
        
//         filtered = append(filtered, traderCandidates...)
//     }
    
//     return filtered, nil
// }

// func (f *BalanceFilter) Priority() int { return 3 }


// // --------------------------------- Система взвешенного отбора

// type WeightedSelector struct {
//     random *rand.Rand
// }

// func (s *WeightedSelector) Select(candidates []*BankDetailView) *BankDetailView {
//     if len(candidates) == 0 {
//         return nil
//     }
//     if len(candidates) == 1 {
//         return candidates[0]
//     }
    
//     // Вычисляем общий вес (сумму приоритетов)
//     totalWeight := 0.0
//     for _, candidate := range candidates {
//         totalWeight += candidate.TraderPriority
//     }
    
//     // Взвешенный случайный выбор
//     r := s.random.Float64() * totalWeight
//     cumulative := 0.0
    
//     for _, candidate := range candidates {
//         cumulative += candidate.TraderPriority
//         if r <= cumulative {
//             return candidate
//         }
//     }
    
//     // Fallback - возвращаем кандидата с максимальным приоритетом
//     return s.selectByMaxPriority(candidates)
// }

// func (s *WeightedSelector) selectByMaxPriority(candidates []*BankDetailView) *BankDetailView {
//     var best *BankDetailView
//     maxPriority := -1.0
    
//     for _, candidate := range candidates {
//         if candidate.TraderPriority > maxPriority {
//             maxPriority = candidate.TraderPriority
//             best = candidate
//         }
//     }
    
//     return best
// }

// //-------------------------------------------- Основной engine каскада
// type CascadeMatchEngine struct {
//     filters []Filter
//     selector *WeightedSelector
//     cache    *BankDetailCache
// }

// func (e *CascadeMatchEngine) FindBestBankDetail(
//     ctx context.Context, 
//     req *MatchRequest,
// ) (*BankDetailView, error) {
    
//     // Сортируем фильтры по приоритету
//     sort.Slice(e.filters, func(i, j int) bool {
//         return e.filters[i].Priority() < e.filters[j].Priority()
//     })
    
//     var candidates []*BankDetailView
    
//     // Последовательно применяем фильтры
//     for _, filter := range e.filters {
//         start := time.Now()
        
//         filtered, err := filter.Filter(ctx, req, candidates)
//         if err != nil {
//             return nil, fmt.Errorf("filter failed: %w", err)
//         }
        
//         candidates = filtered
//         // metrics.FilterDuration.WithLabelValues(filter.Name()).Observe(time.Since(start).Seconds())
        
//         // Ранний выход если кандидатов не осталось
//         if len(candidates) == 0 {
//             // metrics.CascadeAborted.WithLabelValues(filter.Name()).Inc()
//             return nil, nil
//         }
        
//         // metrics.CandidatesAfterFilter.WithLabelValues(filter.Name()).Set(float64(len(candidates)))
//     }
    
//     // Выбор лучшего кандидата
//     if len(candidates) == 0 {
//         return nil, nil
//     }
    
//     return e.selector.Select(candidates), nil
// }

// //--------------------------обновление кэша
// type CacheUpdater struct {
//     db           *gorm.DB
//     cache        *BankDetailCache
//     walletService handlers.HTTPWalletHandler
//     interval     time.Duration
// }

// func (u *CacheUpdater) Start() {
//     ticker := time.NewTicker(u.interval)
//     go func() {
//         for range ticker.C {
//             u.updateCache()
//         }
//     }()
// }

// func (u *CacheUpdater) updateCache() {
//     // Загрузка денормализованных данных одним запросом
//     var bankDetails []*BankDetailView
    
//     err := u.db.Raw(`
//         SELECT 
//             bd.*,
//             t.enabled as traffic_enabled,
//             t.trader_priority,
//             t.trader_reward_percent,
//             t.platform_fee
//         FROM bank_detail_models bd
//         LEFT JOIN traffic_models t ON t.trader_id = bd.trader_id 
//             AND t.merchant_id = ? -- merchant_id будет подставляться динамически
//         WHERE bd.deleted_at IS NULL
//     `, u.getActiveMerchants()).Scan(&bankDetails).Error
    
//     if err != nil {
//         slog.Error("Failed to update bank details cache", "error", err)
//         return
//     }
    
//     // Пакетное обновление балансов
//     u.updateBalances(bankDetails)
    
//     // Обновление in-memory кэша
//     u.cache.Update(bankDetails)
    
//     // metrics.CacheUpdateTime.SetToCurrentTime()
//     // metrics.CacheSize.Set(float64(len(bankDetails)))
// }