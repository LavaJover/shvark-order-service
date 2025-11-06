package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/mappers"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
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

    // // 3. Сохраняем состояние транзакции
    // state := &models.OrderTransactionStateModel{
    //     OrderID:         orderID,
    //     Operation:       operation,
    //     StatusChanged:   true,
    //     WalletProcessed: walletFunc != nil,
    //     CreatedAt:       time.Now(),
    // }
    // if err := tx.Create(state).Error; err != nil {
    //     tx.Rollback()
    //     return fmt.Errorf("failed to save transaction state: %w", err)
    // }

    return tx.Commit().Error
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
    
    safeSortBy := "created_at"
    switch sortBy {
    case "amount_fiat":
        safeSortBy = "amount_fiat"
    case "expires_at":
        safeSortBy = "expires_at"
    case "created_at":
        safeSortBy = "created_at"
    }

    safeSortOrder := "DESC"
    if strings.ToUpper(sortOrder) == "ASC" {
        safeSortOrder = "ASC"
    }

    // УПРОЩЕННЫЙ запрос без JOIN - используем денормализованное поле trader_id
    baseQuery := r.DB.Model(&models.OrderModel{}).
        Where("trader_id = ?", traderID)

    // Применяем фильтры (все на одной таблице!)
    if len(filters.Statuses) > 0 {
        baseQuery = baseQuery.Where("status IN (?)", filters.Statuses)
    }
    
    if filters.MinAmountFiat > 0 {
        baseQuery = baseQuery.Where("amount_fiat >= ?", filters.MinAmountFiat)
    }
    
    if filters.MaxAmountFiat > 0 {
        baseQuery = baseQuery.Where("amount_fiat <= ?", filters.MaxAmountFiat)
    }
    
    if !filters.DateFrom.IsZero() {
        baseQuery = baseQuery.Where("created_at >= ?", filters.DateFrom)
    }
    
    if !filters.DateTo.IsZero() {
        baseQuery = baseQuery.Where("created_at <= ?", filters.DateTo)
    }

    if filters.OrderID != "" {
        baseQuery = baseQuery.Where("id = ?", filters.OrderID)
    }

    if filters.MerchantOrderID != "" {
        baseQuery = baseQuery.Where("merchant_order_id = ?", filters.MerchantOrderID)
    }

    // Подсчет общего количества (быстро, без JOIN'ов)
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

    // Преобразование (BankDetail теперь создается из денормализованных полей)
    orders := make([]*domain.Order, len(orderModels))
    for i, orderModel := range orderModels {
        orders[i] = mappers.ToDomainOrder(&orderModel)
    }

    return orders, total, nil
}

func (r *DefaultOrderRepository) FindExpiredOrders() ([]*domain.Order, error) {
    var orderModels []models.OrderModel
    
    // Простой запрос без JOIN'ов и Preload'ов
    if err := r.DB.
        Where("status = ?", domain.StatusPending).
        Where("expires_at < ?", time.Now()).
        Find(&orderModels).Error; err != nil {
        return nil, err
    }
    
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
) ([]*domain.Order, int64, error) {
    var orders []models.OrderModel
    var total int64

    // Базовый запрос БЕЗ JOIN'ов - работаем только с order_models
    query := r.DB.Model(&models.OrderModel{})

    // Применяем фильтры к денормализованным полям
    if filter.TraderID != "" {
        query = query.Where("trader_id = ?", filter.TraderID)
    }
    if filter.MerchantID != "" {
        query = query.Where("merchant_id = ?", filter.MerchantID)
    }
    if filter.OrderID != "" {
        query = query.Where("id = ?", filter.OrderID)
    }
    if filter.MerchantOrderID != "" {
        query = query.Where("merchant_order_id = ?", filter.MerchantOrderID)
    }
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    if filter.BankCode != "" {
        // Используем денормализованное поле bank_code
        query = query.Where("bank_code = ?", filter.BankCode)
    }
    if !filter.TimeOpeningStart.IsZero() {
        query = query.Where("created_at >= ?", filter.TimeOpeningStart)
    }
    if !filter.TimeOpeningEnd.IsZero() {
        query = query.Where("created_at <= ?", filter.TimeOpeningEnd)
    }
    if filter.AmountFiatMin > 0 {
        query = query.Where("amount_fiat >= ?", filter.AmountFiatMin)
    }
    if filter.AmountFiatMax > 0 {
        query = query.Where("amount_fiat <= ?", filter.AmountFiatMax)
    }
    if filter.Type != "" {
        query = query.Where("type = ?", filter.Type)
    }
    if filter.DeviceID != "" {
        // Используем денормализованное поле device_id
        query = query.Where("device_id = ?", filter.DeviceID)
    }
    if filter.PaymentSystem != "" {
        // Используем денормализованное поле payment_system
        query = query.Where("payment_system = ?", filter.PaymentSystem)
    }

    // Сортировка (безопасная проверка полей)
    safeSort := "created_at DESC" // значение по умолчанию
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
                safeSort = fmt.Sprintf("%s %s", field, order)
            }
        }
    }
    query = query.Order(safeSort)

    // Сначала получаем общее количество записей
    countQuery := query.Session(&gorm.Session{})
    if err := countQuery.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Пагинация и получение данных
    offset := (page - 1) * limit
    err := query.
        Offset(int(offset)).
        Limit(int(limit)).
        Find(&orders).Error
        
    if err != nil {
        return nil, 0, err
    }

    // Преобразуем в доменные объекты (без дополнительных запросов)
    domainOrders := make([]*domain.Order, len(orders))
    for i, order := range orders {
        domainOrders[i] = mappers.ToDomainOrder(&order)
    }

    return domainOrders, total, nil
}


func (r *DefaultOrderRepository) FindPendingOrdersByDeviceID(deviceID string) ([]*domain.Order, error) {
    var orders []models.OrderModel
    
    // Прямой запрос по денормализованному полю device_id
    err := r.DB.
        Where("status = ?", domain.StatusPending).
        Where("device_id = ?", deviceID).
        Where("expires_at > ?", time.Now()).
        Find(&orders).Error
    
    if err != nil {
        return nil, fmt.Errorf("failed to find pending orders: %w", err)
    }

    domainOrders := make([]*domain.Order, len(orders))
    for i, order := range orders {
        domainOrders[i] = mappers.ToDomainOrder(&order)
    }
    
    return domainOrders, nil
}


// Метод для идемпотентности - проверка, не обрабатывалась ли уже сделка
func (r *DefaultOrderRepository) CheckDuplicatePayment(ctx context.Context, orderID string, paymentHash string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.PaymentProcessingLog{}).
		Where("order_id = ? AND payment_hash = ?", orderID, paymentHash).
		Count(&count).Error
	
	return count > 0, err
}

// Логирование обработки платежа для идемпотентности
func (r *DefaultOrderRepository) LogPaymentProcessing(ctx context.Context, log *models.PaymentProcessingLog) error {
	return r.DB.Create(log).Error
}

// SaveAutomaticLog сохраняет доменный объект логa автоматики
func (r *DefaultOrderRepository) SaveAutomaticLog(ctx context.Context, log *domain.AutomaticLog) error {
    if log == nil {
        return fmt.Errorf("automatic log cannot be nil")
    }
    
    // Конвертируем доменный объект в модель
    modelLog := mappers.ToModelAutomaticLog(log)
    
    return r.DB.WithContext(ctx).Create(modelLog).Error
}

// GetAutomaticLogs получает логи автоматики с фильтрацией
func (r *DefaultOrderRepository) GetAutomaticLogs(ctx context.Context, filter *domain.AutomaticLogFilter) ([]*domain.AutomaticLog, error) {
    if filter == nil {
        return nil, fmt.Errorf("filter cannot be nil")
    }
    
    if filter.Limit == 0 {
        filter.Limit = 50 // Дефолтное значение
    }
    
    if filter.Offset < 0 {
        filter.Offset = 0
    }
    
    query := r.DB.WithContext(ctx).Model(&models.AutomaticLogModel{})
    
    // Применяем фильтры
    if filter.DeviceID != "" {
        query = query.Where("device_id = ?", filter.DeviceID)
    }
    
    if filter.TraderID != "" {
        query = query.Where("trader_id = ?", filter.TraderID)
    }
    
    if filter.Success != nil {
        query = query.Where("success = ?", *filter.Success)
    }
    
    if filter.Action != "" {
        query = query.Where("action = ?", filter.Action)
    }
    
    if !filter.StartDate.IsZero() {
        query = query.Where("created_at >= ?", filter.StartDate)
    }
    
    if !filter.EndDate.IsZero() {
        query = query.Where("created_at <= ?", filter.EndDate)
    }
    
    var modelLogs []*models.AutomaticLogModel
    err := query.
        Order("created_at DESC").
        Limit(filter.Limit).
        Offset(filter.Offset).
        Find(&modelLogs).Error
    
    if err != nil {
        return nil, fmt.Errorf("failed to fetch automatic logs: %w", err)
    }
    
    // Конвертируем модели в доменные объекты
    domainLogs := make([]*domain.AutomaticLog, len(modelLogs))
    for i, modelLog := range modelLogs {
        domainLogs[i] = mappers.ToDomainAutomaticLog(modelLog)
    }
    
    return domainLogs, nil
}