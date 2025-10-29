package domain

import (
	"time"

	trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
)

type Traffic struct {
	ID 					string
	MerchantID 			string
	TraderID 			string
	TraderRewardPercent float64
	PlatformFee			float64
	TraderPriority 		float64
	Enabled 			bool // для админов
	Name 				string

	// Гибкие параметры
	ActivityParams 		TrafficActivityParams

	// Для антифрода
	AntifraudParams		TrafficAntifraudParams

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
	SetTraderLockTrafficStatus(traderID string, unlocked bool) error
	SetMerchantLockTrafficStatus(traderID string, unlocked bool) error
	SetManuallyLockTrafficStatus(trafficID string, unlocked bool) error
	SetAntifraudLockTrafficStatus(traderID string, unlocked bool) error
	IsTrafficUnlocked(trafficID string) (*trafficdto.TrafficUnlockedResponse, error)
	GetLockStatuses(trafficID string) (*trafficdto.LockStatusesResponse, error)
	GetTrafficByTraderID(traderID string) ([]*Traffic, error) // НОВОЕ
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
	SetTraderLockTrafficStatus(traderID string, unlocked bool) error
	SetMerchantLockTrafficStatus(traderID string, unlocked bool) error
	SetManuallyLockTrafficStatus(trafficID string, unlocked bool) error
	SetAntifraudLockTrafficStatus(traderID string, unlocked bool) error
	IsTrafficUnlocked(trafficID string) (bool, error)
	GetLockStatuses(trafficID string) (*struct {
		MerchantUnlocked  bool
		TraderUnlocked    bool
		AntifraudUnlocked bool
		ManuallyUnlocked  bool
	}, error)
	GetTrafficByTraderID(traderID string) ([]*Traffic, error) // НОВОЕ
}