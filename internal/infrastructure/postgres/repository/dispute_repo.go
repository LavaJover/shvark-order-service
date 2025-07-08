package repository

import (
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DefaultDisputeRepository struct {
	db *gorm.DB
}

func NewDefaultDisputeRepository(db *gorm.DB) *DefaultDisputeRepository {
	return &DefaultDisputeRepository{db: db}
}

func (r *DefaultDisputeRepository) CreateDispute(dispute *domain.Dispute) error {
	disputeModel := models.DisputeModel{
		ID: uuid.New().String(),
		OrderID: dispute.OrderID,
		ProofUrl: dispute.ProofUrl,
		Reason: dispute.Reason,
		Status: string(domain.DisputeOpened),
		AutoAcceptAt: time.Now().Add(dispute.Ttl),
	}
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

	return &domain.Dispute{
		ID: disputeModel.ID,
		OrderID: disputeModel.OrderID,
		ProofUrl: disputeModel.ProofUrl,
		Reason: disputeModel.Reason,
		Status: domain.DisputeStatus(disputeModel.Status),
	}, nil
}

func (r *DefaultDisputeRepository) GetDisputeByOrderID(orderID string) (*domain.Dispute, error) {
	var disputeModel models.DisputeModel
	if err := r.db.Model(&models.DisputeModel{}).Where("order_id = ?", orderID).First(&disputeModel).Error; err != nil {
		return nil, err
	}

	return &domain.Dispute{
		ID: disputeModel.ID,
		OrderID: disputeModel.OrderID,
		ProofUrl: disputeModel.ProofUrl,
		Reason: disputeModel.Reason,
		Status: domain.DisputeStatus(disputeModel.Status),
	}, nil
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
		disputes[i] = &domain.Dispute{
			ID: disputeModel.ID,
			OrderID: disputeModel.OrderID,
			ProofUrl: disputeModel.ProofUrl,
			Reason: disputeModel.Reason,
			Status: domain.DisputeStatus(disputeModel.Reason),
		}
	}

	return disputes, nil
}