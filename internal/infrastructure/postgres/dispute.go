package postgres

import (
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"gorm.io/gorm"
)

type DisputeModel struct {
	ID 			 string `gorm:"primaryKey"`
	OrderID 	 string	
	ProofUrl 	 string
	Reason 		 string
	Status 		 string
	Order		 OrderModel `gorm:"foreignKey:OrderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CreatedAt	 time.Time
	UpdatedAt 	 time.Time
	AutoAcceptAt time.Time   
}

type DefaultDisputeRepository struct {
	db *gorm.DB
}

func NewDefaultDisputeRepository(db *gorm.DB) *DefaultDisputeRepository {
	return &DefaultDisputeRepository{db: db}
}

func (r *DefaultDisputeRepository) CreateDispute(dispute *domain.Dispute) error {
	disputeModel := DisputeModel{
		OrderID: dispute.OrderID,
		ProofUrl: dispute.ProofUrl,
		Reason: dispute.Reason,
		Status: string(dispute.Status),
		AutoAcceptAt: time.Now().Add(dispute.Ttl),
	}
	if err := r.db.Create(&disputeModel).Error; err != nil {
		return err
	}
	dispute.ID = disputeModel.ID
	return nil
}

func (r *DefaultDisputeRepository) UpdateDisputeStatus(disputeID string, status domain.DisputeStatus) error {
	return r.db.Model(&DisputeModel{ID: disputeID}).Update("status", status).Error
}

func (r *DefaultDisputeRepository) GetDisputeByID(disputeID string) (*domain.Dispute, error) {
	var disputeModel DisputeModel
	if err := r.db.Model(&DisputeModel{}).Where("id = ?", disputeID).First(&disputeModel).Error; err != nil {
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
	var disputeModel DisputeModel
	if err := r.db.Model(&DisputeModel{}).Where("order_id = ?", orderID).First(&disputeModel).Error; err != nil {
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
	var disputeModels []DisputeModel
	if err := r.db.Model(&DisputeModel{}).
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