package repository

import (
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/mappers"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"gorm.io/gorm"
)

type DefaultDisputeRepository struct {
	db *gorm.DB
}

func NewDefaultDisputeRepository(db *gorm.DB) *DefaultDisputeRepository {
	return &DefaultDisputeRepository{db: db}
}

func (r *DefaultDisputeRepository) CreateDispute(dispute *domain.Dispute) error {
	disputeModel := mappers.ToGORMDispute(dispute)
	if err := r.db.Create(&disputeModel).Error; err != nil {
		return err
	}
	dispute.ID = disputeModel.ID
	return nil
}

func (r *DefaultDisputeRepository) UpdateDisputeStatus(disputeID string, status domain.DisputeStatus) error {
	return r.db.Model(&models.DisputeModel{ID: disputeID}).Update("status", status).Error
}

func (r *DefaultDisputeRepository) GetDisputeByID(disputeID string) (*domain.Dispute, error) {
	var disputeModel models.DisputeModel
	if err := r.db.Model(&models.DisputeModel{}).Where("id = ?", disputeID).First(&disputeModel).Error; err != nil {
		return nil, err
	}

	return mappers.ToDomainDispute(&disputeModel), nil
}

func (r *DefaultDisputeRepository) GetDisputeByOrderID(orderID string) (*domain.Dispute, error) {
	var disputeModel models.DisputeModel
	if err := r.db.Model(&models.DisputeModel{}).Where("order_id = ?", orderID).First(&disputeModel).Error; err != nil {
		return nil, err
	}

	return mappers.ToDomainDispute(&disputeModel), nil
}

func (r *DefaultDisputeRepository) FindExpiredDisputes() ([]*domain.Dispute, error) {
	var disputeModels []models.DisputeModel
	if err := r.db.Model(&models.DisputeModel{}).
		Where("status = ?", string(domain.DisputeOpened)).
		Where("auto_accept_at < ?", time.Now()).
		Find(&disputeModels).Error; err != nil {
			return nil, err
		}
	disputes := make([]*domain.Dispute, len(disputeModels))
	for i, disputeModel := range disputeModels {
		disputes[i] = mappers.ToDomainDispute(&disputeModel)
	}

	return disputes, nil
}

func (r *DefaultDisputeRepository) GetOrderDisputes(filter domain.GetDisputesFilter) ([]*domain.Dispute, int64, error) {
    query := r.db.Model(&models.DisputeModel{}).
        Joins("JOIN order_models ON order_models.id = dispute_models.order_id").
        Joins("JOIN bank_detail_models ON bank_detail_models.id = order_models.bank_details_id")
    
    if filter.DisputeID != nil {
        query = query.Where("dispute_models.id = ?", *filter.DisputeID)
    }
    if filter.TraderID != nil {
        query = query.Where("bank_detail_models.trader_id = ?", *filter.TraderID)
    }
    if filter.OrderID != nil {
        query = query.Where("dispute_models.order_id = ?", *filter.OrderID)
    }
    if filter.MerchantID != nil {
        query = query.Where("order_models.merchant_id = ?", *filter.MerchantID)
    }
    if filter.Status != nil {
        query = query.Where("dispute_models.status = ?", *filter.Status)
    }
    
    var total int64
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("count failed: %w", err)
    }
    
    offset := (filter.Page - 1) * filter.Limit
    query = query.Offset(offset).Limit(filter.Limit)
    
    var disputeModels []models.DisputeModel
    if err := query.
        Preload("Order").
        Preload("Order.BankDetail").
        Find(&disputeModels).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to find dispute models: %w", err)
    }
    
    disputes := make([]*domain.Dispute, len(disputeModels))
    for i, disputeModel := range disputeModels {
        disputes[i] = mappers.ToDomainDispute(&disputeModel)
    }
    
    return disputes, total, nil
}