package domain

import "time"

type DisputeStatus string

const (
	DisputeOpened  	DisputeStatus = "DISPUTE_OPENED"
	DisputeAccepted DisputeStatus = "DISPUTE_ACCEPTED"
	DisputeRejected DisputeStatus = "DISPUTE_REJECTED"
	DisputeFreezed  DisputeStatus = "DISPUTE_FREEZED"
)

type Dispute struct {
	ID 			string
	OrderID 	string
	ProofUrl 	string
	Reason 		string
	Status 		DisputeStatus
	Ttl			time.Duration
}

type DisputeRepository interface {
	CreateDispute(dispute *Dispute) error
	UpdateDisputeStatus(disputeID string, status DisputeStatus) error
	GetDisputeByID(disputeID string) (*Dispute, error)
	GetDisputeByOrderID(orderID string) (*Dispute, error)
	FindExpiredDisputes() ([]*Dispute, error)
}

type DisputeUsecase interface {
	CreateDispute(dispute *Dispute) error
	AcceptDispute(disputeID string) error
	RejectDispute(disputeID string) error
	FreezeDispute(disputeID string) error
	GetDisputeByID(disputeID string) (*Dispute, error)
	GetDisputeByOrderID(orderID string) (*Dispute, error)
	AcceptExpiredDisputes() error
}