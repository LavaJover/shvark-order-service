package usecase

import "github.com/LavaJover/shvark-order-service/internal/domain"

type DefaultTrafficUsecase struct {
	TrafficRepo domain.TrafficRepository
}

func NewDefaultTrafficUsecase(trafficRepo domain.TrafficRepository) *DefaultTrafficUsecase {
	return &DefaultTrafficUsecase{TrafficRepo: trafficRepo}
}

func (uc *DefaultTrafficUsecase) AddTraffic(traffic *domain.Traffic) error {
	return uc.TrafficRepo.CreateTraffic(traffic)
}

func (uc *DefaultTrafficUsecase) EditTraffic(traffic *domain.Traffic) error {
	return uc.TrafficRepo.UpdateTraffic(traffic)
}

func (uc *DefaultTrafficUsecase) DeleteTraffic(trafficID string) error {
	return uc.TrafficRepo.DeleteTraffic(trafficID)
}

func (uc *DefaultTrafficUsecase) GetTrafficByID(trafficID string) (*domain.Traffic, error) {
	return uc.TrafficRepo.GetTrafficByID(trafficID)
}

func (uc *DefaultTrafficUsecase) GetTrafficRecords(page, limit int32) ([]*domain.Traffic, error) {
	return uc.TrafficRepo.GetTrafficRecords(page, limit)
}

func (uc *DefaultTrafficUsecase) GetTrafficByTraderMerchant(traderID, merchantID string) (*domain.Traffic, error) {
	return uc.TrafficRepo.GetTrafficByTraderMerchant(traderID, merchantID)
}

func (uc *DefaultTrafficUsecase) DisableTraderTraffic(traderID string) error {
	return uc.TrafficRepo.UpdateTraffic(&domain.Traffic{
		ID: traderID,
		Enabled: false,
	})
}