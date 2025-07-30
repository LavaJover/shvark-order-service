package relationsdto

type CreateTeamRelationInput struct {
	TraderID 		string
	TeamLeadID 		string
	RelationParams 	RelationParams
}

type RelationParams struct {
	Commission float64
}

type UpdateRelationParamsInput struct {
	RelationID 		string
	RelationParams 	RelationParams
}