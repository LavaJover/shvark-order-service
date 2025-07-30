package mappers

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
)

func ToDomainRelationship(model *models.TeamRelationshipModel) *domain.TeamRelationship {
	return &domain.TeamRelationship{
		ID: model.ID,
		TeamLeadID: model.TeamLeadID,
		TraderID: model.TraderID,
		TeamRelationshipRapams: domain.TeamRelationshipRapams{
			Commission: model.Commission,
		},
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func ToGORMRelationship(relation *domain.TeamRelationship) *models.TeamRelationshipModel {
	return &models.TeamRelationshipModel{
		ID: relation.ID,
		TeamLeadID: relation.TeamLeadID,
		TraderID: relation.TraderID,
		Commission: relation.TeamRelationshipRapams.Commission,
		CreatedAt: relation.CreatedAt,
		UpdatedAt: relation.UpdatedAt,
	}
}