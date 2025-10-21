package trafficdto

type LockStatusesResponse struct {
	TrafficID         string `json:"traffic_id"`
	MerchantUnlocked  bool   `json:"merchant_unlocked"`
	TraderUnlocked    bool   `json:"trader_unlocked"`
	AntifraudUnlocked bool   `json:"antifraud_unlocked"`
	ManuallyUnlocked  bool   `json:"manually_unlocked"`
}

type TrafficUnlockedResponse struct {
	TrafficID string `json:"traffic_id"`
	Unlocked  bool   `json:"unlocked"`
}