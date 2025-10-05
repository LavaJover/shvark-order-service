package mappers

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
)

func ToDomainDispute(model *models.DisputeModel) *domain.Dispute {
	return &domain.Dispute{
		ID: model.ID,
		OrderID: model.OrderID,
		DisputeAmountFiat: model.DisputeAmountFiat,
		DisputeAmountCrypto: model.DisputeAmountCrypto,
		DisputeCryptoRate: model.DisputeCryptoRate,
		ProofUrl: model.ProofUrl,
		Reason: model.Reason,
		Status: domain.DisputeStatus(model.Status),
		Ttl: model.Ttl,
		AutoAcceptAt: model.AutoAcceptAt,
		OrderStatusOriginal: domain.OrderStatus(model.OrderStatusOriginal),
		OrderStatusDisputed: domain.OrderStatus(model.OrderStatusDisputed),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func ToGORMDispute(dispute *domain.Dispute) *models.DisputeModel {
	return &models.DisputeModel{
		ID: dispute.ID,
		OrderID: dispute.OrderID,
		DisputeAmountFiat: dispute.DisputeAmountFiat,
		DisputeAmountCrypto: dispute.DisputeAmountCrypto,
		DisputeCryptoRate: dispute.DisputeCryptoRate,
		ProofUrl: dispute.ProofUrl,
		Reason: dispute.Reason,
		Status: string(dispute.Status),
		Ttl: dispute.Ttl,
		AutoAcceptAt: dispute.AutoAcceptAt,
		OrderStatusOriginal: string(dispute.OrderStatusOriginal),
		OrderStatusDisputed: string(dispute.OrderStatusDisputed),
		CreatedAt: dispute.CreatedAt,
		UpdatedAt: dispute.UpdatedAt,
	}
}