package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
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
	orderModel := OrderModel{
		ID: uuid.New().String(),
		MerchantID: order.MerchantID,
		AmountFiat: order.AmountFiat,
		AmountCrypto: order.AmountCrypto,
		Status: domain.StatusCreated,
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
	}

	if err := r.DB.Create(&orderModel).Error; err != nil {
		return "", err
	}

	order.ID = orderModel.ID
	return order.ID, nil
}

func (r *DefaultOrderRepository) GetOrderByID(orderID string) (*domain.Order, error) {
	var order OrderModel
	if err := r.DB.Preload("BankDetail").First(&order, "id = ?", orderID).Error; err != nil {
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
		},
	}, nil
}

func (r *DefaultOrderRepository) UpdateOrderStatus(orderID string, newStatus domain.OrderStatus) error {
	updatedOrderModel := OrderModel{
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
	var orderModels []OrderModel
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
    baseQuery := r.DB.Model(&OrderModel{}).
        Preload("BankDetail").
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
			},
		}
	}

	return orders, total, nil
}

func (r *DefaultOrderRepository) FindExpiredOrders() ([]*domain.Order, error) {
	var orderModels []OrderModel
	if err := r.DB.Preload("BankDetail").
		Where("status = ?", domain.StatusCreated).
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
			},
			ExpiresAt: orderModel.ExpiresAt,
		}
	}

	return orders, nil
}

func (r *DefaultOrderRepository) GetOrdersByBankDetailID(bankDetailID string) ([]*domain.Order, error) {
	var orderModels []OrderModel
	if err := r.DB.
		Preload("BankDetail").
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
			},
			ExpiresAt: orderModel.ExpiresAt,
		}
	}

	return orders, nil
}