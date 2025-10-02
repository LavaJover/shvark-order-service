package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/mappers"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/dto"
	"gorm.io/gorm"
)

type DefaultOrderRepository struct {
	DB *gorm.DB
}

func NewDefaultOrderRepository(db *gorm.DB) *DefaultOrderRepository {
	return &DefaultOrderRepository{DB: db}
}

// ProcessOrderCriticalOperation - выполнение критичной операции в транзакции
func (r *DefaultOrderRepository) ProcessOrderCriticalOperation(
    orderID string, 
    newStatus domain.OrderStatus, 
    operation string, // добавляем параметр операции
    walletFunc func() error,
) error {
    tx := r.DB.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    // 1. Обновляем статус
    if err := tx.Model(&models.OrderModel{}).Where("id = ?", orderID).Update("status", newStatus).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to update order status: %w", err)
    }

    // 2. Выполняем операцию с кошельком
    if walletFunc != nil {
        if err := walletFunc(); err != nil {
            tx.Rollback()
            return fmt.Errorf("wallet operation failed: %w", err)
        }
    }

    // 3. Сохраняем состояние транзакции
    state := &models.OrderTransactionStateModel{
        OrderID:         orderID,
        Operation:       operation,
        StatusChanged:   true,
        WalletProcessed: walletFunc != nil,
        CreatedAt:       time.Now(),
    }
    if err := tx.Create(state).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to save transaction state: %w", err)
    }

    return tx.Commit().Error
}

// GetTransactionState - получение состояния транзакции
func (r *DefaultOrderRepository) GetTransactionState(orderID string) (*domain.OrderTransactionStateModel, error) {
    var state models.OrderTransactionStateModel
    err := r.DB.Where("order_id = ?", orderID).
        Order("created_at DESC").
        First(&state).Error
    if err != nil {
        return nil, err
    }
    return &domain.OrderTransactionStateModel{
        ID: state.ID,
        OrderID: state.OrderID,
        Operation: state.Operation,
        StatusChanged: state.StatusChanged,
        WalletProcessed: state.WalletProcessed,
        EventPublished: state.EventPublished,
        CallbackSent: state.CallbackSent,
        CreatedAt: state.CreatedAt,
        CompletedAt: state.CompletedAt,
        UpdatedAt: state.UpdatedAt,
    }, nil
}

// UpdateTransactionState - обновление состояния транзакции
func (r *DefaultOrderRepository) UpdateTransactionState(orderID string, updates map[string]interface{}) error {
    return r.DB.Model(&models.OrderTransactionStateModel{}).
        Where("order_id = ?", orderID).
        Order("created_at DESC").
        Limit(1).
        Updates(updates).Error
}

// MarkEventPublished - отметка успешной публикации события
func (r *DefaultOrderRepository) MarkEventPublished(orderID string) error {
    return r.UpdateTransactionState(orderID, map[string]interface{}{
        "event_published": true,
        "updated_at":      time.Now(),
    })
}

// MarkCallbackSent - отметка успешной отправки callback
func (r *DefaultOrderRepository) MarkCallbackSent(orderID string) error {
    return r.UpdateTransactionState(orderID, map[string]interface{}{
        "callback_sent": true,
        "updated_at":    time.Now(),
    })
}

// MarkCompleted - отметка завершения всех операций
func (r *DefaultOrderRepository) MarkCompleted(orderID string) error {
    now := time.Now()
    return r.UpdateTransactionState(orderID, map[string]interface{}{
        "completed_at": &now,
        "updated_at":   now,
    })
}

// FindInconsistentOrders - упрощенная версия поиска несоответствий
// FindInconsistentOrders - исправленная версия с приведением типов
func (r *DefaultOrderRepository) FindInconsistentOrders() ([]string, error) {
    var inconsistentOrderIDs []string

    // Исправляем JOIN с явным приведением типов
    query := `
        SELECT DISTINCT ots.order_id
        FROM order_transaction_states ots
        JOIN order_models o ON ots.order_id = o.id::text  -- Приводим UUID к TEXT
        WHERE ots.status_changed = true 
        AND ots.wallet_processed = false
        AND ots.created_at < $1
        AND (
            (o.status = $2 AND o.released_at IS NULL) OR  -- CANCELED но не разморожено
            (o.status = $3 AND o.released_at IS NULL) OR  -- COMPLETED но не освобождено  
            (o.status = $4 AND ots.operation = 'create')  -- PENDING но заморозка не прошла
        )
    `

    if err := r.DB.Raw(query, 
        time.Now().Add(-5*time.Minute), // $1
        domain.StatusCanceled,          // $2
        domain.StatusCompleted,         // $3
        domain.StatusPending,           // $4
    ).Pluck("order_id", &inconsistentOrderIDs).Error; err != nil {
        return nil, fmt.Errorf("failed to find inconsistent orders: %w", err)
    }

    return inconsistentOrderIDs, nil
}

// GetInconsistentOrderDetails - получение деталей несоответствия для заказа
func (r *DefaultOrderRepository) GetInconsistentOrderDetails(orderID string) (map[string]interface{}, error) {
    // Получаем информацию о заказе
    var order models.OrderModel
    if err := r.DB.Where("id = ?", orderID).First(&order).Error; err != nil {
        return nil, err
    }

    // Получаем состояние транзакции
    var state models.OrderTransactionStateModel
    if err := r.DB.Where("order_id = ?", orderID).
        Order("created_at DESC").First(&state).Error; err != nil {
        return nil, err
    }

    details := map[string]interface{}{
        "order_id":          orderID,
        "status":           order.Status,
        "released_at":      order.ReleasedAt,
        "status_changed":   state.StatusChanged,
        "wallet_processed": state.WalletProcessed,
        "operation":        state.Operation,
        "created_at":       state.CreatedAt,
    }

    return details, nil
}

func (r *DefaultOrderRepository) CreateOrder(order *domain.Order) error {
	orderModel := mappers.ToGORMOrder(order)
	if err := r.DB.Create(orderModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *DefaultOrderRepository) GetOrderByID(orderID string) (*domain.Order, error) {
	var order models.OrderModel
	if err := r.DB.Preload("BankDetail", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped() // отключаем фильтрацию по DeletedAt
	}).First(&order, "id = ?", orderID).Error; err != nil {
		return nil, err
	}

	return mappers.ToDomainOrder(&order), nil
}

func (r *DefaultOrderRepository) GetOrderByMerchantOrderID(merchantOrderID string) (*domain.Order, error) {
	var order models.OrderModel
	if err := r.DB.Preload("BankDetail", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped() // отключаем фильтрацию по DeletedAt
	}).First(&order, "merchant_order_id = ?", merchantOrderID).Error; err != nil {
		return nil, err
	}

	return mappers.ToDomainOrder(&order), nil
}

func (r *DefaultOrderRepository) UpdateOrderStatus(orderID string, newStatus domain.OrderStatus) error {
	updatedOrderModel := models.OrderModel{
		ID: orderID,
		Status: newStatus,
	}

	if err := r.DB.Updates(&updatedOrderModel).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultOrderRepository) GetOrdersByTraderID(
	traderID string, 
	page, limit int64, 
	sortBy, sortOrder string,
	filters domain.OrderFilters,
) ([]*domain.Order, int64, error) {
	var orderModels []models.OrderModel
	var total int64
	
	safeSortBy := "order_models.created_at"
	switch sortBy {
	case "amount_fiat":
		safeSortBy = "order_models.amount_fiat"
	case "expires_at":
		safeSortBy = "order_models.expires_at"
	case "created_at":
		safeSortBy = "order_models.created_at"
	}

	safeSortOrder := "DESC"
	if strings.ToUpper(sortOrder) == "ASC" {
		safeSortOrder = "ASC"
	}

    // Базовый запрос с JOIN
    baseQuery := r.DB.Model(&models.OrderModel{}).
	Preload("BankDetail", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped() // отключаем фильтрацию по DeletedAt
	}).
        Joins("JOIN bank_detail_models ON order_models.bank_details_id = bank_detail_models.id").
        Where("bank_detail_models.trader_id = ?", traderID)

    // Применяем фильтры
    if len(filters.Statuses) > 0 {
        baseQuery = baseQuery.Where("order_models.status IN (?)", filters.Statuses)
    }
    
    if filters.MinAmountFiat > 0 {
        baseQuery = baseQuery.Where("order_models.amount_fiat >= ?", filters.MinAmountFiat)
    }
    
    if filters.MaxAmountFiat > 0 {
        baseQuery = baseQuery.Where("order_models.amount_fiat <= ?", filters.MaxAmountFiat)
    }
    
    if !filters.DateFrom.IsZero() {
        baseQuery = baseQuery.Where("order_models.created_at >= ?", filters.DateFrom)
    }
    
    if !filters.DateTo.IsZero() {
        baseQuery = baseQuery.Where("order_models.created_at <= ?", filters.DateTo)
    }

	if filters.OrderID != "" {
		baseQuery = baseQuery.Where("order_models.id = ?", filters.OrderID)
	}

	if filters.MerchantOrderID != "" {
		baseQuery = baseQuery.Where("order_models.merchant_order_id = ?", filters.MerchantOrderID)
	}

    // Подсчет общего количества с учетом фильтров
    if err := baseQuery.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count orders: %w", err)
    }

    // Основной запрос с сортировкой и пагинацией
    offset := (page - 1) * limit
    err := baseQuery.
        Order(fmt.Sprintf("%s %s", safeSortBy, safeSortOrder)).
        Offset(int(offset)).
        Limit(int(limit)).
        Find(&orderModels).Error

    if err != nil {
        return nil, 0, fmt.Errorf("failed to find orders: %w", err)
    }

	orders := make([]*domain.Order, len(orderModels))
	for i, orderModel := range orderModels {
		orders[i] = mappers.ToDomainOrder(&orderModel)
	}

	return orders, total, nil
}

func (r *DefaultOrderRepository) FindExpiredOrders() ([]*domain.Order, error) {
	var orderModels []models.OrderModel
	if err := r.DB.Preload("BankDetail", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped() // отключаем фильтрацию по DeletedAt
	}).
		Where("status = ?", domain.StatusPending).
		Where("expires_at < ?", time.Now()).
		Find(&orderModels).Error; err != nil {return nil, err}
	
	orders := make([]*domain.Order, len(orderModels))
	for i, orderModel := range orderModels {
		orders[i] = mappers.ToDomainOrder(&orderModel)
	}

	return orders, nil
}

func (r *DefaultOrderRepository) GetOrdersByBankDetailID(bankDetailID string) ([]*domain.Order, error) {
	var orderModels []models.OrderModel
	if err := r.DB.
	Preload("BankDetail", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped() // отключаем фильтрацию по DeletedAt
	}).
		Where("bank_details_id = ?", bankDetailID).
		Find(&orderModels).Error; 
		err != nil {
			return nil, err
		}

	orders := make([]*domain.Order, len(orderModels))
	for i, orderModel := range orderModels {
		orders[i] = mappers.ToDomainOrder(&orderModel)
	}

	return orders, nil
}

func (r *DefaultOrderRepository) GetCreatedOrdersByClientID(clientID string) ([]*domain.Order, error) {
	var orderModels []models.OrderModel
	if err := r.DB.Model(&models.OrderModel{}).Preload("BankDetail", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped() // отключаем фильтрацию по DeletedAt
	}).Where("client_id = ? AND status = ?", clientID, domain.StatusPending).Find(&orderModels).Error; err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, len(orderModels))
	for i, orderModel := range orderModels {
		orders[i] = mappers.ToDomainOrder(&orderModel)
	}

	return orders, nil
}

func (r *DefaultOrderRepository) GetOrderStatistics(traderID string, dateFrom, dateTo time.Time) (*domain.OrderStatistics, error) {
	var stats domain.OrderStatistics

	baseQuery := func() *gorm.DB {
		return r.DB.
			Model(&models.OrderModel{}).
			Joins("JOIN bank_detail_models AS bdm ON bdm.id = order_models.bank_details_id").
			Where("bdm.trader_id = ?", traderID).
			Where("order_models.created_at BETWEEN ? AND ?", dateFrom, dateTo)
	}

	// Всего сделок
	if err := baseQuery().Count(&stats.TotalOrders).Error; err != nil {
		return nil, fmt.Errorf("count total orders: %w", err)
	}

	// Succeed сделки
	type SucceedAgg struct {
		Count     int64
		SumFiat   float64
		SumCrypto float64
		Income    float64
	}
	var succ SucceedAgg
	if err := baseQuery().
		Where("order_models.status = ?", domain.StatusCompleted).
		Select("COUNT(*) as count, SUM(amount_fiat) as sum_fiat, SUM(amount_crypto) as sum_crypto, SUM(amount_crypto * trader_reward_percent) as income").
		Scan(&succ).Error; err != nil {
		return nil, fmt.Errorf("succeed agg: %w", err)
	}

	stats.SucceedOrders = succ.Count
	stats.ProcessedAmountFiat = succ.SumFiat
	stats.ProcessedAmountCrypto = succ.SumCrypto
	stats.IncomeCrypto = succ.Income

	// Canceled сделки
	type CancelAgg struct {
		Count     int64
		SumFiat   float64
		SumCrypto float64
	}
	var canc CancelAgg
	if err := baseQuery().
		Where("order_models.status = ?", domain.StatusCanceled).
		Select("COUNT(*) as count, SUM(amount_fiat) as sum_fiat, SUM(amount_crypto) as sum_crypto").
		Scan(&canc).Error; err != nil {
		return nil, fmt.Errorf("canceled agg: %w", err)
	}

	stats.CanceledOrders = canc.Count
	stats.CanceledAmountFiat = canc.SumFiat
	stats.CanceledAmountCrypto = canc.SumCrypto

	return &stats, nil
}

func (r *DefaultOrderRepository) GetOrders(filter domain.Filter, sortField string, page, size int) ([]*domain.Order, int64, error) {
    query := r.DB.Model(&models.OrderModel{}).Preload("BankDetail")

    // Применяем фильтры
    if filter.DealID != nil {
        // ИСПРАВЛЕНО: используем правильное имя поля
        query = query.Where("merchant_order_id = ?", *filter.DealID)
    }
    if filter.Type != nil {
        query = query.Where("type = ?", *filter.Type)
    }
    if filter.Status != nil {
        query = query.Where("status = ?", *filter.Status)
    }
    if filter.TimeOpeningStart != nil {
        query = query.Where("created_at >= ?", *filter.TimeOpeningStart)
    }
    if filter.TimeOpeningEnd != nil {
        query = query.Where("created_at <= ?", *filter.TimeOpeningEnd)
    }
    if filter.AmountMin != nil {
        query = query.Where("amount_fiat >= ?", *filter.AmountMin)
    }
    if filter.AmountMax != nil {
        query = query.Where("amount_fiat <= ?", *filter.AmountMax)
    }

    query = query.Where("merchant_id = ?", filter.MerchantID)

    // Считаем общее количество
    var total int64
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("count failed: %w", err)
    }

    // Применяем сортировку и пагинацию только если нужно
    if sortField != "" {
        // ИСПРАВЛЕНО: проверяем на пустую строку
        mappedField := MapSortField(sortField)
        query = query.Order(fmt.Sprintf("%s DESC", mappedField))
    }
    
    if size > 0 {
        offset := page * size
        query = query.Offset(offset).Limit(size)
    }

    // Выполняем запрос
    var orderModels []models.OrderModel
    if err := query.Find(&orderModels).Error; err != nil {
        return nil, 0, fmt.Errorf("find failed: %w", err)
    }

    orders := make([]*domain.Order, len(orderModels))
    for i, orderModel := range orderModels {
        orders[i] = mappers.ToDomainOrder(&orderModel)
    }

    return orders, total, nil
}

// MapSortField без изменений
func MapSortField(input string) string {
    switch input {
    case "id":
        return "id"
    case "deal_id":
        return "merchant_order_id"
    case "time_opening":
        return "created_at"
    case "time_expires":
        return "expires_at"
    case "time_complete":
        return "updated_at"
    case "type":
        return "type"
    case "status":
        return "status"
    case "amount":
        return "amount_fiat"
    default:
        return "created_at"
    }
}

func (r *DefaultOrderRepository) GetAllOrders(
    filter *domain.AllOrdersFilters, 
    sort string,
    page, limit int32,
) ([]*domain.Order, int64, error) { // Добавляем возврат общего количества
    var orders []models.OrderModel
    var total int64

    // Базовый запрос с JOIN банковских данных
    query := r.DB.Model(&models.OrderModel{}).
        Joins("JOIN bank_detail_models ON bank_detail_models.id = order_models.bank_details_id")

    // Применяем фильтры
    if filter.TraderID != "" {
        query = query.Where("bank_detail_models.trader_id = ?", filter.TraderID)
    }
    if filter.MerchantID != "" {
        query = query.Where("order_models.merchant_id = ?", filter.MerchantID)
    }
    if filter.OrderID != "" {
        query = query.Where("order_models.id = ?", filter.OrderID)
    }
    if filter.MerchantOrderID != "" {
        query = query.Where("order_models.merchant_order_id = ?", filter.MerchantOrderID)
    }
    if filter.Status != "" {
        query = query.Where("order_models.status = ?", filter.Status)
    }
    if filter.BankCode != "" {
        query = query.Where("bank_detail_models.bank_code = ?", filter.BankCode)
    }
    if !filter.TimeOpeningStart.IsZero() {
        query = query.Where("order_models.created_at >= ?", filter.TimeOpeningStart)
    }
    if !filter.TimeOpeningEnd.IsZero() {
        query = query.Where("order_models.created_at <= ?", filter.TimeOpeningEnd)
    }
    if filter.AmountFiatMin > 0 {
        query = query.Where("order_models.amount_fiat >= ?", filter.AmountFiatMin)
    }
    if filter.AmountFiatMax > 0 {
        query = query.Where("order_models.amount_fiat <= ?", filter.AmountFiatMax)
    }
    if filter.Type != "" {
        query = query.Where("order_models.type = ?", filter.Type)
    }
    if filter.DeviceID != "" {
        query = query.Where("bank_detail_models.device_id = ?", filter.DeviceID)
    }
    if filter.PaymentSystem != "" {
        query = query.Where("bank_detail_models.payment_system = ?", filter.PaymentSystem)
    }

    // Сортировка (безопасная проверка полей)
    safeSort := "order_models.created_at DESC" // значение по умолчанию
    if sort != "" {
        allowedSorts := map[string]bool{
            "amount_fiat": true,
            "created_at":  true,
            "expires_at":  true,
        }
        
        sortParts := strings.Split(sort, " ")
        if len(sortParts) == 2 {
            field := strings.ToLower(sortParts[0])
            order := strings.ToUpper(sortParts[1])
            
            if allowedSorts[field] && (order == "ASC" || order == "DESC") {
                safeSort = fmt.Sprintf("order_models.%s %s", field, order)
            }
        }
    }
    query = query.Order(safeSort)

    // Пагинация
    offset := (page - 1) * limit
    err := query.
        Offset(int(offset)).
        Limit(int(limit)).
        Preload("BankDetail"). // Подгружаем связанные банковские данные
        Find(&orders).Error
        
    if err != nil {
        return nil, 0, err
    }

    // Получаем общее количество записей (без лимита)
    if err := query.Offset(-1).Limit(-1).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Преобразуем в доменные объекты
    domainOrders := make([]*domain.Order, len(orders))
    for i, order := range orders {
        domainOrders[i] = mappers.ToDomainOrder(&order)
    }

    return domainOrders, total, nil
}

// CancelExpiredOrdersBatch - батчевая отмена просроченных заказов
func (r *DefaultOrderRepository) CancelExpiredOrdersBatch(ctx context.Context) ([]dto.ExpiredOrderData, error) {
    query := `
        WITH expired_orders AS (
            SELECT o.id, o.merchant_id, o.amount_fiat, o.currency, o.callback_url, 
                   o.merchant_order_id, o.trader_reward_percent, o.platform_fee,
                   bd.trader_id, bd.bank_name, bd.phone, bd.card_number, bd.owner
            FROM order_models o
            JOIN bank_detail_models bd ON o.bank_details_id = bd.id
            WHERE o.status = ? AND o.expires_at < NOW()
            FOR UPDATE SKIP LOCKED
        ),
        updated AS (
            UPDATE order_models 
            SET status = ?, updated_at = NOW()
            WHERE id IN (SELECT id FROM expired_orders)
            RETURNING id
        )
        SELECT eo.* FROM expired_orders eo
        JOIN updated u ON eo.id = u.id
    `

    var expiredOrders []dto.ExpiredOrderData

    if err := r.DB.WithContext(ctx).Raw(query, 
        domain.StatusPending, 
        domain.StatusCanceled,
    ).Scan(&expiredOrders).Error; err != nil {
        return nil, err
    }

    return expiredOrders, nil
}

// IncrementReleaseAttempts - увеличение счетчика попыток разморозки
func (r *DefaultOrderRepository) IncrementReleaseAttempts(ctx context.Context, orderIDs []string) error {
    return r.DB.WithContext(ctx).
        Model(&models.OrderModel{}).
        Where("id IN ?", orderIDs).
        UpdateColumn("release_attempts", gorm.Expr("release_attempts + 1")).
        Error
}

// IncrementCallbackAttempts - увеличение счетчика попыток callback'ов
func (r *DefaultOrderRepository) IncrementCallbackAttempts(ctx context.Context, orderIDs []string) error {
    if len(orderIDs) == 0 {
        return nil
    }
    
    return r.DB.WithContext(ctx).
        Model(&models.OrderModel{}).
        Where("id IN ?", orderIDs).
        UpdateColumn("callback_attempts", gorm.Expr("callback_attempts + 1")).
        Error
}

// MarkCallbacksSentAt - отметка успешной отправки callback'ов
func (r *DefaultOrderRepository) MarkCallbacksSentAt(ctx context.Context, orderIDs []string) error {
    if len(orderIDs) == 0 {
        return nil
    }
    
    return r.DB.WithContext(ctx).
        Model(&models.OrderModel{}).
        Where("id IN ?", orderIDs).
        Updates(map[string]interface{}{
            "callbacks_sent_at": time.Now(),
        }).Error
}

// LoadExpiredOrderDataByIDs - загрузка данных expired orders по списку ID
func (r *DefaultOrderRepository) LoadExpiredOrderDataByIDs(ctx context.Context, orderIDs []string) ([]dto.ExpiredOrderData, error) {
    if len(orderIDs) == 0 {
        return []dto.ExpiredOrderData{}, nil
    }

    query := `
        SELECT o.id, o.merchant_id, o.amount_fiat, o.currency, o.callback_url, 
               o.merchant_order_id, o.trader_reward_percent, o.platform_fee,
               bd.trader_id, bd.bank_name, bd.phone, bd.card_number, bd.owner
        FROM order_models o
        JOIN bank_detail_models bd ON o.bank_details_id = bd.id
        WHERE o.id IN ?
        AND o.status = ?
    `

    var expiredOrders []dto.ExpiredOrderData

    if err := r.DB.WithContext(ctx).Raw(query, orderIDs, domain.StatusCanceled).Scan(&expiredOrders).Error; err != nil {
        return nil, fmt.Errorf("failed to load expired order data by IDs: %w", err)
    }

    return expiredOrders, nil
}

// MarkReleasedAt - отметка успешной разморозки
func (r *DefaultOrderRepository) MarkReleasedAt(ctx context.Context, orderIDs []string) error {
    return r.DB.WithContext(ctx).
        Model(&models.OrderModel{}).
        Where("id IN ?", orderIDs).
        Updates(map[string]interface{}{
            "released_at": time.Now(),
        }).Error
}

// IncrementPublishAttempts - увеличение счетчика попыток публикации
func (r *DefaultOrderRepository) IncrementPublishAttempts(ctx context.Context, orderIDs []string) error {
    if len(orderIDs) == 0 {
        return nil
    }
    
    return r.DB.WithContext(ctx).
        Model(&models.OrderModel{}).
        Where("id IN ?", orderIDs).
        UpdateColumn("publish_attempts", gorm.Expr("publish_attempts + 1")).
        Error
}

// MarkPublishedAt - отметка успешной публикации
func (r *DefaultOrderRepository) MarkPublishedAt(ctx context.Context, orderIDs []string) error {
    if len(orderIDs) == 0 {
        return nil
    }
    
    return r.DB.WithContext(ctx).
        Model(&models.OrderModel{}).
        Where("id IN ?", orderIDs).
        Updates(map[string]interface{}{
            "published_at": time.Now(),
        }).Error
}


// FindStuckOrders - поиск "зависших" ордеров (отменены, но не разморожены)
func (r *DefaultOrderRepository) FindStuckOrders(ctx context.Context, maxAttempts int) ([]string, error) {
    var orderIDs []string
    err := r.DB.WithContext(ctx).
        Model(&models.OrderModel{}).
        Select("id").
        Where("status = ? AND released_at IS NULL AND release_attempts < ?", 
              domain.StatusCanceled, maxAttempts).
        Where("updated_at < ?", time.Now().Add(-10*time.Minute)). // старше 10 минут
        Pluck("id", &orderIDs).Error
    return orderIDs, err
}
