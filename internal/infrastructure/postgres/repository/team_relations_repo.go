package repository

import (
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/mappers"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"github.com/jaevor/go-nanoid"
	"gorm.io/gorm"
)

var nanoIDGenerator, _ = nanoid.CustomASCII("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", 12)

type DefaultTeamRelationsRepository struct {
	DB *gorm.DB
}

func NewDefaultTeamRelationsRepository(db *gorm.DB) *DefaultTeamRelationsRepository{
	return &DefaultTeamRelationsRepository{
		DB: db,
	}
}

func (r *DefaultTeamRelationsRepository) CreateRelationship(relation *domain.TeamRelationship) error {
	id := nanoIDGenerator()
	relation.ID = id
	model := mappers.ToGORMRelationship(relation)
	return r.DB.Create(model).Error
}

func (r *DefaultTeamRelationsRepository) GetRelationshipsByTeamLeadID(teamLeadID string) ([]*domain.TeamRelationship, error) {
	var models []models.TeamRelationshipModel
	if err := r.DB.
		Where("team_lead_id = ?", teamLeadID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	relationships := make([]*domain.TeamRelationship, len(models))
	for i, model := range models {
		relationships[i] = mappers.ToDomainRelationship(&model)
	}
	return relationships, nil
}

func (r *DefaultTeamRelationsRepository) UpdateRelationshipParams(relationship *domain.TeamRelationship) error {
	model := mappers.ToGORMRelationship(relationship)
	return r.DB.Model(model).
		Updates(map[string]interface{}{
			"commission": model.Commission,
			"updated_at": time.Now(),
		}).Error
}

func (r *DefaultTeamRelationsRepository) GetRelationshipsByTraderID(traderID string) ([]*domain.TeamRelationship, error) {
	var models []models.TeamRelationshipModel
	if err := r.DB.
		Where("trader_id = ?", traderID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	relationships := make([]*domain.TeamRelationship, len(models))
	for i, model := range models {
		relationships[i] = mappers.ToDomainRelationship(&model)
	}
	return relationships, nil
}

func (r *DefaultTeamRelationsRepository) DeleteRelationship(relationID string) error {
	return r.DB.Delete(models.TeamRelationshipModel{ID: relationID}).Error
}