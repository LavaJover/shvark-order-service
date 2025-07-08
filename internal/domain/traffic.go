package domain

type Traffic struct {
	ID 					string
	MerchantID 			string
	TraderID 			string
	TraderRewardPercent float64
	PlatformFee			float64
	TraderPriority 		float64
	Enabled 			bool
}

type TrafficUsecase interface {
	AddTraffic(traffic *Traffic) error
	EditTraffic(traffic *Traffic) error
	GetTrafficRecords(page, limit int32) ([]*Traffic, error)
	GetTrafficByID(trafficID string) (*Traffic, error)
	DeleteTraffic(trafficID string) error
	GetTrafficByTraderMerchant(traderID, merchantID string) (*Traffic, error)
	DisableTraderTraffic(traderID string) error
	EnableTraderTraffic(traderID string) error
	GetTraderTrafficStatus(traderID string) (bool, error)
}

type TrafficRepository interface {
	CreateTraffic(traffic *Traffic) error
	UpdateTraffic(traffic *Traffic) error
	GetTrafficRecords(page, limit int32) ([]*Traffic, error)
	GetTrafficByID(trafficID string) (*Traffic, error)
	DeleteTraffic(trafficID string) error
	GetTrafficByTraderMerchant(traderID, merchantID string) (*Traffic, error)
	DisableTraderTraffic(traderID string) error
	EnableTraderTraffic(traderID string) error
	GetTraderTrafficStatus(traderID string) (bool, error)
}