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
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type GetDisputesFilter struct {
	DisputeID 	*string
	TraderID  	*string
	OrderID   	*string
	MerchantID 	*string
	Status 		*string
	Page 		int
	Limit 		int
}

type DisputeRepository interface {
	CreateDispute(dispute *Dispute) error
	UpdateDisputeStatus(disputeID string, status DisputeStatus) error
	GetDisputeByID(disputeID string) (*Dispute, error)
	GetDisputeByOrderID(orderID string) (*Dispute, error)
	FindExpiredDisputes() ([]*Dispute, error)
	GetOrderDisputes(filter GetDisputesFilter) ([]*Dispute, int64, error)
}