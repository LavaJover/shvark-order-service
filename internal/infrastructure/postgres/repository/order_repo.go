package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DefaultOrderRepository struct {
	DB *gorm.DB
}

func NewDefaultOrderRepository(db *gorm.DB) *DefaultOrderRepository {
	return &DefaultOrderRepository{DB: db}
}

func (r *DefaultOrderRepository) CreateOrder(order *domain.Order) (string, error) {
	orderModel := models.OrderModel{
		ID: uuid.New().String(),
		MerchantID: order.MerchantID,
		AmountFiat: order.AmountFiat,
		AmountCrypto: order.AmountCrypto,
		Status: domain.StatusPending,
		Currency: order.Currency,
		Country: order.Country,
		ClientID: order.ClientID,
		PaymentSystem: order.PaymentSystem,
		BankDetailsID: order.BankDetailsID,
		ExpiresAt: order.ExpiresAt,
		MerchantOrderID: order.MerchantOrderID,
		Shuffle: order.Shuffle,
		CallbackURL: order.CallbackURL,
		TraderRewardPercent: order.TraderRewardPercent,
		Recalculated: order.Recalculated,
		CryptoRubRate: order.CryptoRubRate,
		PlatformFee: order.PlatformFee,
		Type: order.Type,
	}

	if err := r.DB.Create(&orderModel).Error; err != nil {
		return "", err
	}

	order.ID = orderModel.ID
	return order.ID, nil
}

func (r *DefaultOrderRepository) GetOrderByID(orderID string) (*domain.Order, error) {
	var order models.OrderModel
	if err := r.DB.Preload("BankDetail", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped() // отключаем фильтрацию по DeletedAt
	}).First(&order, "id = ?", orderID).Error; err != nil {
		return nil, err
	}

	return &domain.Order{
		ID: order.ID,
		MerchantID: order.MerchantID,
		AmountFiat: order.AmountFiat,
		AmountCrypto: order.AmountCrypto,
		Currency: order.Currency,
		Country: order.Country,
		ClientID: order.ClientID,
		Status: order.Status,
		PaymentSystem: order.PaymentSystem,
		BankDetailsID: order.BankDetailsID,
		ExpiresAt: order.ExpiresAt,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
		MerchantOrderID: order.MerchantOrderID,
		Shuffle: order.Shuffle,
		CallbackURL: order.CallbackURL,
		TraderRewardPercent: order.TraderRewardPercent,
		Recalculated: order.Recalculated,
		CryptoRubRate: order.CryptoRubRate,
		PlatformFee: order.PlatformFee,
		Type: order.Type,
		BankDetail: &domain.BankDetail{
			ID: order.BankDetail.ID,
			TraderID: order.BankDetail.TraderID,
			Country: order.BankDetail.Country,
			Currency: order.BankDetail.Currency,
			MinAmount: order.BankDetail.MinAmount,
			MaxAmount: order.BankDetail.MaxAmount,
			BankName: order.BankDetail.BankName,
			PaymentSystem: order.BankDetail.PaymentSystem,
			Delay: order.BankDetail.Delay,
			Enabled: order.BankDetail.Enabled,
			CardNumber: order.BankDetail.CardNumber,
			Phone: order.BankDetail.Phone,
			Owner: order.BankDetail.Owner,
			MaxOrdersSimultaneosly: order.BankDetail.MaxOrdersSimultaneosly,
			MaxAmountDay: order.BankDetail.MaxAmountDay,
			MaxAmountMonth: order.BankDetail.MaxAmountMonth,
			InflowCurrency: order.BankDetail.InflowCurrency,
			BankCode: order.BankDetail.BankCode,
			NspkCode: order.BankDetail.NspkCode,
			DeviceID: order.BankDetail.DeviceID,
			MaxQuantityDay: order.BankDetail.MaxQuantityDay,
			MaxQuantityMonth: order.BankDetail.MaxQuantityMonth,
			CreatedAt: order.BankDetail.CreatedAt,
			UpdatedAt: order.BankDetail.UpdatedAt,
		},
	}, nil
}

func (r *DefaultOrderRepository) GetOrderByMerchantOrderID(merchantOrderID string) (*domain.Order, error) {
	var order models.OrderModel
	if err := r.DB.Preload("BankDetail", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped() // отключаем фильтрацию по DeletedAt
	}).First(&order, "merchant_order_id = ?", merchantOrderID).Error; err != nil {
		return nil, err
	}

	return &domain.Order{
		ID: order.ID,
		MerchantID: order.MerchantID,
		AmountFiat: order.AmountFiat,
		AmountCrypto: order.AmountCrypto,
		Currency: order.Currency,
		Country: order.Country,
		ClientID: order.ClientID,
		Status: order.Status,
		PaymentSystem: order.PaymentSystem,
		BankDetailsID: order.BankDetailsID,
		ExpiresAt: order.ExpiresAt,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
		MerchantOrderID: order.MerchantOrderID,
		Shuffle: order.Shuffle,
		CallbackURL: order.CallbackURL,
		TraderRewardPercent: order.TraderRewardPercent,
		Recalculated: order.Recalculated,
		CryptoRubRate: order.CryptoRubRate,
		PlatformFee: order.PlatformFee,
		Type: order.Type,
		BankDetail: &domain.BankDetail{
			ID: order.BankDetail.ID,
			TraderID: order.BankDetail.TraderID,
			Country: order.BankDetail.Country,
			Currency: order.BankDetail.Currency,
			MinAmount: order.BankDetail.MinAmount,
			MaxAmount: order.BankDetail.MaxAmount,
			BankName: order.BankDetail.BankName,
			PaymentSystem: order.BankDetail.PaymentSystem,
			Delay: order.BankDetail.Delay,
			Enabled: order.BankDetail.Enabled,
			CardNumber: order.BankDetail.CardNumber,
			Phone: order.BankDetail.Phone,
			Owner: order.BankDetail.Owner,
			MaxOrdersSimultaneosly: order.BankDetail.MaxOrdersSimultaneosly,
			MaxAmountDay: order.BankDetail.MaxAmountDay,
			MaxAmountMonth: order.BankDetail.MaxAmountMonth,
			InflowCurrency: order.BankDetail.InflowCurrency,
			BankCode: order.BankDetail.BankCode,
			NspkCode: order.BankDetail.NspkCode,
			DeviceID: order.BankDetail.DeviceID,
			MaxQuantityDay: order.BankDetail.MaxQuantityDay,
			MaxQuantityMonth: order.BankDetail.MaxQuantityMonth,
			CreatedAt: order.BankDetail.CreatedAt,
			UpdatedAt: order.BankDetail.UpdatedAt,
		},
	}, nil
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
		orders[i] = &domain.Order{
			ID: orderModel.ID,
			MerchantID: orderModel.MerchantID,
			AmountFiat: orderModel.AmountFiat,
			AmountCrypto: orderModel.AmountCrypto,
			Currency: orderModel.Currency,
			Country: orderModel.Country,
			ClientID: orderModel.ClientID,
			Status: orderModel.Status,
			PaymentSystem: orderModel.PaymentSystem,
			ExpiresAt: orderModel.ExpiresAt,
			BankDetailsID: orderModel.BankDetailsID,
			CreatedAt: orderModel.CreatedAt,
			UpdatedAt: orderModel.UpdatedAt,
			MerchantOrderID: orderModel.MerchantOrderID,
			Shuffle: orderModel.Shuffle,
			CallbackURL: orderModel.CallbackURL,
			TraderRewardPercent: orderModel.TraderRewardPercent,
			Recalculated: orderModel.Recalculated,
			CryptoRubRate: orderModel.CryptoRubRate,
			PlatformFee: orderModel.PlatformFee,
			Type: orderModel.Type,
			BankDetail: &domain.BankDetail{
				ID: orderModel.BankDetail.ID,
				TraderID: orderModel.BankDetail.TraderID,
				Country: orderModel.BankDetail.Country,
				Currency: orderModel.BankDetail.Currency,
				MinAmount: orderModel.BankDetail.MinAmount,
				MaxAmount: orderModel.BankDetail.MaxAmount,
				BankName: orderModel.BankDetail.BankName,
				PaymentSystem: orderModel.BankDetail.PaymentSystem,
				Delay: orderModel.BankDetail.Delay,
				Enabled: orderModel.BankDetail.Enabled,
				CardNumber: orderModel.BankDetail.CardNumber,
				Phone: orderModel.BankDetail.Phone,
				Owner: orderModel.BankDetail.Owner,
				MaxOrdersSimultaneosly: orderModel.BankDetail.MaxOrdersSimultaneosly,
				MaxAmountDay: orderModel.BankDetail.MaxAmountDay,
				MaxAmountMonth: orderModel.BankDetail.MaxAmountMonth,
				InflowCurrency: orderModel.BankDetail.InflowCurrency,
				BankCode: orderModel.BankDetail.BankCode,
				NspkCode: orderModel.BankDetail.NspkCode,
				MaxQuantityDay: orderModel.BankDetail.MaxQuantityDay,
				MaxQuantityMonth: orderModel.BankDetail.MaxQuantityMonth,
				DeviceID: orderModel.BankDetail.DeviceID,
				CreatedAt: orderModel.BankDetail.CreatedAt,
				UpdatedAt: orderModel.BankDetail.UpdatedAt,
			},
		}
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
		orders[i] = &domain.Order{
			ID: orderModel.ID,
			MerchantID: orderModel.MerchantID,
			AmountFiat: orderModel.AmountFiat,
			AmountCrypto: orderModel.AmountCrypto,
			Currency: orderModel.Currency,
			Country: orderModel.Country,
			ClientID: orderModel.ClientID,
			Status: orderModel.Status,
			PaymentSystem: orderModel.PaymentSystem,
			BankDetailsID: orderModel.BankDetailsID,
			CreatedAt: orderModel.CreatedAt,
			UpdatedAt: orderModel.UpdatedAt,
			MerchantOrderID: orderModel.MerchantOrderID,
			Shuffle: orderModel.Shuffle,
			CallbackURL: orderModel.CallbackURL,
			TraderRewardPercent: orderModel.TraderRewardPercent,
			Recalculated: orderModel.Recalculated,
			CryptoRubRate: orderModel.CryptoRubRate,
			PlatformFee: orderModel.PlatformFee,
			Type: orderModel.Type,
			BankDetail: &domain.BankDetail{
				ID: orderModel.BankDetail.ID,
				TraderID: orderModel.BankDetail.TraderID,
				Country: orderModel.BankDetail.Country,
				Currency: orderModel.BankDetail.Currency,
				MinAmount: orderModel.BankDetail.MinAmount,
				MaxAmount: orderModel.BankDetail.MaxAmount,
				BankName: orderModel.BankDetail.BankName,
				PaymentSystem: orderModel.BankDetail.PaymentSystem,
				Delay: orderModel.BankDetail.Delay,
				Enabled: orderModel.BankDetail.Enabled,
				CardNumber: orderModel.BankDetail.CardNumber,
				Phone: orderModel.BankDetail.Phone,
				Owner: orderModel.BankDetail.Owner,
				MaxOrdersSimultaneosly: orderModel.BankDetail.MaxOrdersSimultaneosly,
				MaxAmountDay: orderModel.BankDetail.MaxAmountDay,
				MaxAmountMonth: orderModel.BankDetail.MaxAmountMonth,
				InflowCurrency: orderModel.BankDetail.InflowCurrency,
				BankCode: orderModel.BankDetail.BankCode,
				NspkCode: orderModel.BankDetail.NspkCode,
				MaxQuantityDay: orderModel.BankDetail.MaxQuantityDay,
				MaxQuantityMonth: orderModel.BankDetail.MaxQuantityMonth,
				DeviceID: orderModel.BankDetail.DeviceID,
				CreatedAt: orderModel.BankDetail.CreatedAt,
				UpdatedAt: orderModel.BankDetail.UpdatedAt,
			},
			ExpiresAt: orderModel.ExpiresAt,
		}
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
		orders[i] = &domain.Order{
			ID: orderModel.ID,
			MerchantID: orderModel.MerchantID,
			AmountFiat: orderModel.AmountFiat,
			AmountCrypto: orderModel.AmountCrypto,
			Currency: orderModel.Currency,
			Country: orderModel.Country,
			ClientID: orderModel.ClientID,
			Status: orderModel.Status,
			PaymentSystem: orderModel.PaymentSystem,
			BankDetailsID: orderModel.BankDetailsID,
			CreatedAt: orderModel.CreatedAt,
			UpdatedAt: orderModel.UpdatedAt,
			MerchantOrderID: orderModel.MerchantOrderID,
			Shuffle: orderModel.Shuffle,
			CallbackURL: orderModel.CallbackURL,
			TraderRewardPercent: orderModel.TraderRewardPercent,
			Recalculated: orderModel.Recalculated,
			CryptoRubRate: orderModel.CryptoRubRate,
			PlatformFee: orderModel.PlatformFee,
			Type: orderModel.Type,
			BankDetail: &domain.BankDetail{
				ID: orderModel.BankDetail.ID,
				TraderID: orderModel.BankDetail.TraderID,
				Country: orderModel.BankDetail.Country,
				Currency: orderModel.BankDetail.Currency,
				MinAmount: orderModel.BankDetail.MinAmount,
				MaxAmount: orderModel.BankDetail.MaxAmount,
				BankName: orderModel.BankDetail.BankName,
				PaymentSystem: orderModel.BankDetail.PaymentSystem,
				Delay: orderModel.BankDetail.Delay,
				Enabled: orderModel.BankDetail.Enabled,
				CardNumber: orderModel.BankDetail.CardNumber,
				Phone: orderModel.BankDetail.Phone,
				Owner: orderModel.BankDetail.Owner,
				MaxOrdersSimultaneosly: orderModel.BankDetail.MaxOrdersSimultaneosly,
				MaxAmountDay: orderModel.BankDetail.MaxAmountDay,
				MaxAmountMonth: orderModel.BankDetail.MaxAmountMonth,
				InflowCurrency: orderModel.BankDetail.InflowCurrency,
				BankCode: orderModel.BankDetail.BankCode,
				NspkCode: orderModel.BankDetail.NspkCode,
				MaxQuantityDay: orderModel.BankDetail.MaxQuantityDay,
				MaxQuantityMonth: orderModel.BankDetail.MaxQuantityMonth,
				DeviceID: orderModel.BankDetail.DeviceID,
				CreatedAt: orderModel.BankDetail.CreatedAt,
				UpdatedAt: orderModel.BankDetail.UpdatedAt,
			},
			ExpiresAt: orderModel.ExpiresAt,
		}
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
		orders[i] = &domain.Order{
			ID: orderModel.ID,
			MerchantID: orderModel.MerchantID,
			AmountFiat: orderModel.AmountFiat,
			AmountCrypto: orderModel.AmountCrypto,
			Currency: orderModel.Currency,
			Country: orderModel.Country,
			ClientID: orderModel.ClientID,
			Status: orderModel.Status,
			PaymentSystem: orderModel.PaymentSystem,
			BankDetailsID: orderModel.BankDetailsID,
			CreatedAt: orderModel.CreatedAt,
			UpdatedAt: orderModel.UpdatedAt,
			MerchantOrderID: orderModel.MerchantOrderID,
			Shuffle: orderModel.Shuffle,
			CallbackURL: orderModel.CallbackURL,
			TraderRewardPercent: orderModel.TraderRewardPercent,
			Recalculated: orderModel.Recalculated,
			CryptoRubRate: orderModel.CryptoRubRate,
			PlatformFee: orderModel.PlatformFee,
			Type: orderModel.Type,
			BankDetail: &domain.BankDetail{
				ID: orderModel.BankDetail.ID,
				TraderID: orderModel.BankDetail.TraderID,
				Country: orderModel.BankDetail.Country,
				Currency: orderModel.BankDetail.Currency,
				MinAmount: orderModel.BankDetail.MinAmount,
				MaxAmount: orderModel.BankDetail.MaxAmount,
				BankName: orderModel.BankDetail.BankName,
				PaymentSystem: orderModel.BankDetail.PaymentSystem,
				Delay: orderModel.BankDetail.Delay,
				Enabled: orderModel.BankDetail.Enabled,
				CardNumber: orderModel.BankDetail.CardNumber,
				Phone: orderModel.BankDetail.Phone,
				Owner: orderModel.BankDetail.Owner,
				MaxOrdersSimultaneosly: orderModel.BankDetail.MaxOrdersSimultaneosly,
				MaxAmountDay: orderModel.BankDetail.MaxAmountDay,
				MaxAmountMonth: orderModel.BankDetail.MaxAmountMonth,
				InflowCurrency: orderModel.BankDetail.InflowCurrency,
				BankCode: orderModel.BankDetail.BankCode,
				NspkCode: orderModel.BankDetail.NspkCode,
				MaxQuantityDay: orderModel.BankDetail.MaxQuantityDay,
				MaxQuantityMonth: orderModel.BankDetail.MaxQuantityMonth,
				DeviceID: orderModel.BankDetail.DeviceID,
				CreatedAt: orderModel.BankDetail.CreatedAt,
				UpdatedAt: orderModel.BankDetail.UpdatedAt,
			},
			ExpiresAt: orderModel.ExpiresAt,
		}
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
        orders[i] = &domain.Order{
            ID:                 orderModel.ID,
            MerchantID:         orderModel.MerchantID,
            AmountFiat:         orderModel.AmountFiat,
            AmountCrypto:       orderModel.AmountCrypto,
            Currency:           orderModel.Currency,
            Country:            orderModel.Country,
            ClientID:           orderModel.ClientID,
            Status:             orderModel.Status,
            PaymentSystem:      orderModel.PaymentSystem,
            BankDetailsID:      orderModel.BankDetailsID,
            CreatedAt:          orderModel.CreatedAt,
            UpdatedAt:          orderModel.UpdatedAt,
            MerchantOrderID:    orderModel.MerchantOrderID,
            Shuffle:            orderModel.Shuffle,
            CallbackURL:        orderModel.CallbackURL,
            TraderRewardPercent: orderModel.TraderRewardPercent,
            Recalculated:       orderModel.Recalculated,
            CryptoRubRate:      orderModel.CryptoRubRate,
            PlatformFee:        orderModel.PlatformFee,
            Type:               orderModel.Type,
            BankDetail: &domain.BankDetail{
                ID:          orderModel.BankDetail.ID,
                TraderID:    orderModel.BankDetail.TraderID,
                Country:     orderModel.BankDetail.Country,
                Currency:    orderModel.BankDetail.Currency,
                MinAmount:   orderModel.BankDetail.MinAmount,
                MaxAmount:   orderModel.BankDetail.MaxAmount,
                BankName:    orderModel.BankDetail.BankName,
                PaymentSystem: orderModel.BankDetail.PaymentSystem,
                Delay:       orderModel.BankDetail.Delay,
                Enabled:     orderModel.BankDetail.Enabled,
                CardNumber:  orderModel.BankDetail.CardNumber,
                Phone:       orderModel.BankDetail.Phone,
                Owner:       orderModel.BankDetail.Owner,
                MaxOrdersSimultaneosly: orderModel.BankDetail.MaxOrdersSimultaneosly,
                MaxAmountDay:           orderModel.BankDetail.MaxAmountDay,
                MaxAmountMonth:         orderModel.BankDetail.MaxAmountMonth,
                InflowCurrency:         orderModel.BankDetail.InflowCurrency,
                BankCode:               orderModel.BankDetail.BankCode,
                NspkCode:               orderModel.BankDetail.NspkCode,
                MaxQuantityDay:         orderModel.BankDetail.MaxQuantityDay,
                MaxQuantityMonth:       orderModel.BankDetail.MaxQuantityMonth,
                DeviceID:               orderModel.BankDetail.DeviceID,
                CreatedAt:              orderModel.BankDetail.CreatedAt,
                UpdatedAt:              orderModel.BankDetail.UpdatedAt,
            },
            ExpiresAt: orderModel.ExpiresAt,
        }
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