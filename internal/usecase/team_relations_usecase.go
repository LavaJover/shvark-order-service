package usecase

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	relationsdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/relations"
)

type TeamRelationsUsecase interface {
	GetRelationshipsByTeamLeadID(teamLeadID string) ([]*domain.TeamRelationship, error)
	CreateRelationship(input *relationsdto.CreateTeamRelationInput) error
	UpdateRelationshipParams(input *relationsdto.UpdateRelationParamsInput) error
	GetRelationshipsByTraderID(traderID string) ([]*domain.TeamRelationship, error)
	DeleteTeamRelationship(relationID string) error
}

type DefaultTeamRelationsUsecase struct {
	teamRelationsRepo domain.TeamRelationRepository
}

func NewDefaultTeamRelationsUsecase(repo domain.TeamRelationRepository) *DefaultTeamRelationsUsecase {
	return &DefaultTeamRelationsUsecase{
		teamRelationsRepo: repo,
	}
}

func (uc *DefaultTeamRelationsUsecase) CreateRelationship(input *relationsdto.CreateTeamRelationInput) error {
	return uc.teamRelationsRepo.CreateRelationship(
		&domain.TeamRelationship{
			TeamLeadID: input.TeamLeadID,
			TraderID: input.TraderID,
			TeamRelationshipRapams: domain.TeamRelationshipRapams{
				Commission: input.RelationParams.Commission,
			},
		},
	)
}

func (uc *DefaultTeamRelationsUsecase) UpdateRelationshipParams(input *relationsdto.UpdateRelationParamsInput) error {
	return uc.teamRelationsRepo.UpdateRelationshipParams(
		&domain.TeamRelationship{
			ID: input.RelationID,
			TeamRelationshipRapams: domain.TeamRelationshipRapams{
				Commission: input.RelationParams.Commission,
			},
		},
	)
}

func (uc *DefaultTeamRelationsUsecase) GetRelationshipsByTeamLeadID(teamLeadID string) ([]*domain.TeamRelationship, error) {
	return uc.teamRelationsRepo.GetRelationshipsByTeamLeadID(teamLeadID)
}

func (uc *DefaultTeamRelationsUsecase) GetRelationshipsByTraderID(traderID string) ([]*domain.TeamRelationship, error) {
	return uc.teamRelationsRepo.GetRelationshipsByTraderID(traderID)
}

func (uc *DefaultTeamRelationsUsecase) DeleteTeamRelationship(relationID string) error {
	return uc.teamRelationsRepo.DeleteRelationship(relationID)
}