package postgres

import (
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
		ClientEmail: order.ClientEmail,
		MetadataJSON: order.MetadataJSON,
		PaymentSystem: order.PaymentSystem,
		BankDetailsID: order.BankDetailsID,
		ExpiresAt: order.ExpiresAt,
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
		ClientEmail: order.ClientEmail,
		MetadataJSON: order.MetadataJSON,
		Status: order.Status,
		PaymentSystem: order.PaymentSystem,
		BankDetailsID: order.BankDetailsID,
		ExpiresAt: order.ExpiresAt,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
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

func (r *DefaultOrderRepository) GetOrdersByTraderID(traderID string) ([]*domain.Order, error) {
	var orderModels []OrderModel

	if err := r.DB.
		Preload("BankDetail").
		Joins("JOIN bank_detail_models ON order_models.bank_details_id = bank_detail_models.id").
		Where("bank_detail_models.trader_id = ?", traderID).
		Find(&orderModels).Error; err != nil {
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
			ClientEmail: orderModel.ClientEmail,
			MetadataJSON: orderModel.MetadataJSON,
			Status: orderModel.Status,
			PaymentSystem: orderModel.PaymentSystem,
			ExpiresAt: orderModel.ExpiresAt,
			BankDetailsID: orderModel.BankDetailsID,
			CreatedAt: orderModel.CreatedAt,
			UpdatedAt: orderModel.UpdatedAt,
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

	return orders, nil
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
			ClientEmail: orderModel.ClientEmail,
			MetadataJSON: orderModel.MetadataJSON,
			Status: orderModel.Status,
			PaymentSystem: orderModel.PaymentSystem,
			BankDetailsID: orderModel.BankDetailsID,
			CreatedAt: orderModel.CreatedAt,
			UpdatedAt: orderModel.UpdatedAt,
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
			ClientEmail: orderModel.ClientEmail,
			MetadataJSON: orderModel.MetadataJSON,
			Status: orderModel.Status,
			PaymentSystem: orderModel.PaymentSystem,
			BankDetailsID: orderModel.BankDetailsID,
			CreatedAt: orderModel.CreatedAt,
			UpdatedAt: orderModel.UpdatedAt,
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