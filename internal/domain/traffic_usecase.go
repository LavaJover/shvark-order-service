package domain

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