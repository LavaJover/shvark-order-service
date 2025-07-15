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
	ID 				  	string
	OrderID 		  	string
	DisputeAmountFiat 	float64
	DisputeAmountCrypto float64
	DisputeCryptoRate 	float64
	ProofUrl 			string
	Reason 				string
	Status 				DisputeStatus
	Ttl					time.Duration
	AutoAcceptAt		time.Time
}

type DisputeRepository interface {
	CreateDispute(dispute *Dispute) error
	UpdateDisputeStatus(disputeID string, status DisputeStatus) error
	GetDisputeByID(disputeID string) (*Dispute, error)
	GetDisputeByOrderID(orderID string) (*Dispute, error)
	FindExpiredDisputes() ([]*Dispute, error)
	GetOrderDisputes(page, limit int64, status string) ([]*Dispute, int64, error)
}

type DisputeUsecase interface {
	CreateDispute(dispute *Dispute) error
	AcceptDispute(disputeID string) error
	RejectDispute(disputeID string) error
	FreezeDispute(disputeID string) error
	GetDisputeByID(disputeID string) (*Dispute, error)
	GetDisputeByOrderID(orderID string) (*Dispute, error)
	AcceptExpiredDisputes() error
	GetOrderDisputes(page, limit int64, status string) ([]*Dispute, int64, error)
}