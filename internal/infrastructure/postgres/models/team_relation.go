package models

import "time"

type TeamRelationshipModel struct{
	ID 				string  `gorm:"primaryKey"`
	TeamLeadID 		string 	`gorm:"not null"`
	TraderID 		string	`gotm:"not null"`
	Commission  	float64 
	RelationStart 	time.Time
	RelationEnd 	time.Time
	CreatedAt   	time.Time
	UpdatedAt   	time.Time
}