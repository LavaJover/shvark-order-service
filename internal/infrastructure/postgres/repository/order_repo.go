package repository

import (
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