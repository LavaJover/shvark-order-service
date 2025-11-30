package grpcapi

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/usecase"
	relationsdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/relations"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen/order"
)

type TeamRelationsHandler struct{
	teamRelationsUc usecase.TeamRelationsUsecase
	orderpb.UnimplementedTeamRelationsServiceServer
}

func NewTeamRelationsHandler(
	teamRelationUc usecase.TeamRelationsUsecase,
) *TeamRelationsHandler {
	return &TeamRelationsHandler{
		teamRelationsUc: teamRelationUc,
	}
}

func (h *TeamRelationsHandler) CreateTeamRelation(ctx context.Context, r *orderpb.CreateTeamRelationRequest) (*orderpb.CreateTeamRelationResponse, error) {
	err := h.teamRelationsUc.CreateRelationship(
		&relationsdto.CreateTeamRelationInput{
			TraderID: r.TraderId,
			TeamLeadID: r.TeamLeadId,
			RelationParams: relationsdto.RelationParams{
				Commission: r.Commission,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return &orderpb.CreateTeamRelationResponse{}, nil
}

func (h *TeamRelationsHandler) GetRelationsByTeamLeadID(ctx context.Context, r *orderpb.GetRelationsByTeamLeadIDRequest) (*orderpb.GetRelationsByTeamLeadIDResponse, error) {
	relations, err := h.teamRelationsUc.GetRelationshipsByTeamLeadID(r.TeamLeadId)
	if err != nil {
		return nil, err
	}

	respRelations := make([]*orderpb.TeamRelationship, len(relations))
	for i, relation := range relations {
		respRelations[i] = &orderpb.TeamRelationship{
			Id: relation.ID,
			TeamLeadId: relation.TeamLeadID,
			TraderId: relation.TraderID,
			Commission: relation.TeamRelationshipRapams.Commission,
		}
	}

	return &orderpb.GetRelationsByTeamLeadIDResponse{
		TeamRelations: respRelations,
	}, nil
}

func (h *TeamRelationsHandler) UpdateRelationParams(ctx context.Context, r *orderpb.UpdateRelationParamsRequest) (*orderpb.UpdateRelationParamsResponse, error) {
	err := h.teamRelationsUc.UpdateRelationshipParams(
		&relationsdto.UpdateRelationParamsInput{
			RelationID: r.Relation.Id,
			RelationParams: relationsdto.RelationParams{
				Commission: r.Relation.Commission,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return &orderpb.UpdateRelationParamsResponse{}, nil
}

func (h *TeamRelationsHandler) DeleteTeamRelationship(ctx context.Context, r *orderpb.DeleteTeamRelationshipRequest) (*orderpb.DeleteTeamRelationshipResponse, error) {
	err := h.teamRelationsUc.DeleteTeamRelationship(r.RelationId)
	if err != nil {
		return nil, err
	}

	return &orderpb.DeleteTeamRelationshipResponse{}, nil
}