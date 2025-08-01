package repository

import (
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

func (r *DefaultDisputeRepository) GetOrderDisputes(page, limit int64, status string) ([]*domain.Dispute, int64, error) {
	var (
		disputeModels []models.DisputeModel
		total         int64
	)

	query := r.db.Model(&models.DisputeModel{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(limit)).Find(&disputeModels).Error; err != nil {
		return nil, 0, err
	}

	disputes := make([]*domain.Dispute, len(disputeModels))
	for i, dm := range disputeModels {
		disputes[i] = mappers.ToDomainDispute(&dm)
	}

	return disputes, total, nil
}