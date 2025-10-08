package usecase

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
)

type DefaultTrafficUsecase struct {
	TrafficRepo domain.TrafficRepository
}

func NewDefaultTrafficUsecase(trafficRepo domain.TrafficRepository) *DefaultTrafficUsecase {
	return &DefaultTrafficUsecase{TrafficRepo: trafficRepo}
}

func (uc *DefaultTrafficUsecase) AddTraffic(traffic *domain.Traffic) error {
	return uc.TrafficRepo.CreateTraffic(traffic)
}

func (uc *DefaultTrafficUsecase) EditTraffic(input *trafficdto.EditTrafficInput) error {
	return uc.TrafficRepo.UpdateTraffic(input)
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
	return uc.TrafficRepo.DisableTraderTraffic(traderID)
}

func (uc *DefaultTrafficUsecase) EnableTraderTraffic(traderID string) error {
	return uc.TrafficRepo.EnableTraderTraffic(traderID)
}

func (uc *DefaultTrafficUsecase) GetTraderTrafficStatus(traderID string) (bool, error) {
	return uc.TrafficRepo.GetTraderTrafficStatus(traderID)
}