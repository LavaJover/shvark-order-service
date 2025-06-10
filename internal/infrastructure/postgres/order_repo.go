package postgres

import (
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
		Amount: order.Amount,
		Status: domain.StatusCreated,
		Currency: order.Currency,
		Country: order.Country,
		ClientEmail: order.ClientEmail,
		MetadataJSON: order.MetadataJSON,
		PaymentSystem: order.PaymentSystem,
		BankDetailsID: order.BankDetailsID,
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
		Amount: order.Amount,
		Currency: order.Currency,
		Country: order.Country,
		ClientEmail: order.ClientEmail,
		MetadataJSON: order.MetadataJSON,
		Status: order.Status,
		PaymentSystem: order.PaymentSystem,
		BankDetailsID: order.BankDetailsID,
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

func (r *DefaultOrderRepository) UpdateOrderStatus(orderID string, newStatus string) error {
	updatedOrderModel := OrderModel{
		ID: orderID,
		Status: domain.OrderStatus(newStatus),
	}

	if err := r.DB.Updates(&updatedOrderModel).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultOrderRepository) GetOrdersByTraderID(traderID string) ([]*domain.Order, error) {
	var orderModels []OrderModel

	if err := r.DB.Preload("BankDetail").Where("trader_id = ?", traderID).Find(&orderModels).Error; err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, len(orderModels))
	for i, orderModel := range orderModels {
		orders[i] = &domain.Order{
			ID: orderModel.ID,
			MerchantID: orderModel.MerchantID,
			Amount: orderModel.Amount,
			Currency: orderModel.Currency,
			Country: orderModel.Country,
			ClientEmail: orderModel.ClientEmail,
			MetadataJSON: orderModel.MetadataJSON,
			Status: orderModel.Status,
			PaymentSystem: orderModel.PaymentSystem,
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