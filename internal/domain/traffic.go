package domain

import (
	"time"

	trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
)

// message ActivityParameters {
// 	bool merchant_unlocked = 1;
// 	bool trader_unlocked = 2;
// 	bool manually_unlocked = 3;
// 	bool antifraud_unlocked = 4;

// 	message AntifraudParameters {
// 		bool antifraud_enabled = 1;
// 	}
// }

// message BusinessParameters {
// 	google.protobuf.Duration merchant_deals_duration = 1;
// }

type Traffic struct {
	ID 					string
	MerchantID 			string
	TraderID 			string
	TraderRewardPercent float64
	PlatformFee			float64
	TraderPriority 		float64
	Enabled 			bool // для админов

	// Гибкие параметры
	ActivityParams 		TrafficActivityParams

	// Для антифрода
	AntifraudParams		TrafficAntifraudParams

	// Детали блокировки
	LockDetails			*TrafficLockDetails

	// Бизнес-параметры
	BusinessParams		TrafficBusinessParams
}

type TrafficActivityParams struct {
	MerchantUnlocked	bool
	TraderUnlocked		bool
	AntifraudUnlocked	bool
	ManuallyUnlocked	bool
}

type TrafficAntifraudParams struct {
	AntifraudRequired bool
}

type TrafficLockDetails struct {
	LockedAt			time.Time
	UnlockedAt			time.Time
	Reason				string
}

type TrafficBusinessParams struct {
	MerchantDealsDuration time.Duration
}

type TrafficUsecase interface {
	AddTraffic(traffic *Traffic) error
	EditTraffic(input *trafficdto.EditTrafficInput) error
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
	UpdateTraffic(input *trafficdto.EditTrafficInput) error
	GetTrafficRecords(page, limit int32) ([]*Traffic, error)
	GetTrafficByID(trafficID string) (*Traffic, error)
	DeleteTraffic(trafficID string) error
	GetTrafficByTraderMerchant(traderID, merchantID string) (*Traffic, error)
	DisableTraderTraffic(traderID string) error
	EnableTraderTraffic(traderID string) error
	GetTraderTrafficStatus(traderID string) (bool, error)
}