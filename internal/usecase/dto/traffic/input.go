package trafficdto

import "time"

type EditTrafficInput struct {
	ID 						string
	StoreID              	*string  // Добавляем поле для смены стора 					
	MerchantID 				*string 				
	TraderID 				*string 				
	TraderReward 			*float64 				
	TraderPriority 			*float64 				
	PlatformFee 			*float64				
	Enabled 				*bool
	Name					*string 					
	ActivityParams 			*TrafficActivityParams 	
	AntifraudParams 		*TrafficAntifraudParams 
	BusinessParams 			*TrafficBusinessParams 	
}

type TrafficActivityParams struct {
	TraderUnlocked   	bool 
	AntifraudUnlocked 	bool 
	ManuallyUnlocked  	bool 
}

type TrafficAntifraudParams struct {
	AntifraudRequired bool 
}

type TrafficBusinessParams struct {
	MerchantDealsDuration time.Duration
}