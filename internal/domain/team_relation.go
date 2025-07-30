package domain

import "time"

type TeamRelationship struct{
	ID 					   string
	TeamLeadID 			   string
	TraderID 			   string
	TeamRelationshipRapams TeamRelationshipRapams
	CreatedAt   		   time.Time
	UpdatedAt   		   time.Time
}

type TeamRelationshipRapams struct {
	Commission float64
}

type TeamRelationRepository interface {
	GetRelationshipsByTeamLeadID(teamLeadID string) ([]*TeamRelationship, error)
	CreateRelationship(relationship *TeamRelationship) error
	UpdateRelationshipParams(relationship *TeamRelationship) error
	GetRelationshipsByTraderID(traderID string) ([]*TeamRelationship, error)
	DeleteRelationship(relationID string) error
}