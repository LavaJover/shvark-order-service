package postgres

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"gorm.io/gorm"
	"github.com/google/uuid"
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
		Status: order.Status,
		Currency: order.Currency,
		Country: order.Country,
		ClientEmail: order.ClientEmail,
		MetadataJSON: order.MetadataJSON,
	}

	if err := r.DB.Create(&orderModel).Error; err != nil {
		return "", err
	}

	order.ID = orderModel.ID
	return order.ID, nil
}