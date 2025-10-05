package grpcapi

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/protobuf/types/known/durationpb"
)

type TrafficHandler struct {
	orderpb.UnimplementedTrafficServiceServer
	trafficUsecase domain.TrafficUsecase
}

func NewTrafficHandler(trafficUsecase domain.TrafficUsecase) *TrafficHandler {
	return &TrafficHandler{trafficUsecase: trafficUsecase}
}

func (h *TrafficHandler) AddTraffic(ctx context.Context, r *orderpb.AddTrafficRequest) (*orderpb.AddTrafficResponse, error) {
	traffic := &domain.Traffic{
		MerchantID: r.MerchantId,
		TraderID: r.TraderId,
		TraderRewardPercent: r.TraderRewardPercent,
		TraderPriority: r.TraderPriority,
		Enabled: r.Enabled,
		PlatformFee: r.PlatformFee,
		ActivityParams: domain.TrafficActivityParams{
			MerchantUnlocked: r.ActivityParams.MerchantUnlocked,
			TraderUnlocked: r.ActivityParams.TraderUnlocked,
			AntifraudUnlocked: r.ActivityParams.AntifraudUnlocked,
			ManuallyUnlocked: r.ActivityParams.ManuallyUnlocked,
		},
		AntifraudParams: domain.TrafficAntifraudParams{
			AntifraudRequired: r.AntifraudParams.AntifraudRequired,
		},
		BusinessParams: domain.TrafficBusinessParams{
			MerchantDealsDuration: r.BusinessParams.MerchantDealsDuration.AsDuration(),
		},
	}

	if err := h.trafficUsecase.AddTraffic(traffic); err != nil {
		return &orderpb.AddTrafficResponse{
			Message: "failed to add new traffic",
		}, err
	}

	return &orderpb.AddTrafficResponse{
		Message: "Successfully added new traffic",
	}, nil
}

func (h *TrafficHandler) EditTraffic(ctx context.Context, r *orderpb.EditTrafficRequest) (*orderpb.EditTrafficResponse, error) {
	traffic := &domain.Traffic{
		ID: r.Traffic.Id,
		MerchantID: r.Traffic.MerchantId,
		TraderID: r.Traffic.TraderId,
		TraderRewardPercent: r.Traffic.TraderRewardPercent,
		TraderPriority: r.Traffic.TraderPriority,
		PlatformFee: r.Traffic.PlatformFee,
		Enabled: r.Traffic.Enabled,
	}

	if err := h.trafficUsecase.EditTraffic(traffic); err != nil {
		return &orderpb.EditTrafficResponse{
			Message: "failed to update traffic",
		}, nil
	}

	return &orderpb.EditTrafficResponse{
		Message: "traffic updated successfully",
	}, nil
}

func (h *TrafficHandler) DeleteTraffic(ctx context.Context, r *orderpb.DeleteTrafficRequest) (*orderpb.DeleteTrafficResponse, error) {
	trafficID := r.TrafficId
	if err := h.trafficUsecase.DeleteTraffic(trafficID); err != nil {
		return &orderpb.DeleteTrafficResponse{
			Message: "Failed to delete traffic",
		}, err
	}

	return &orderpb.DeleteTrafficResponse{
		Message: "Deleted traffic successfully",
	}, nil
}

func (h *TrafficHandler) GetTrafficRecords(ctx context.Context, r *orderpb.GetTrafficRecordsRequest) (*orderpb.GetTrafficRecordsResponse, error) {
	page, limit := r.Page, r.Limit
	trafficRecords, err := h.trafficUsecase.GetTrafficRecords(page, limit)
	if err != nil {
		return nil, err
	}

	trafficResponse := make([]*orderpb.Traffic, len(trafficRecords))
	for i, trafficRecord := range trafficRecords {
		trafficResponse[i] = &orderpb.Traffic{
			Id: trafficRecord.ID,
			MerchantId: trafficRecord.MerchantID,
			TraderId: trafficRecord.TraderID,
			TraderRewardPercent: trafficRecord.TraderRewardPercent,
			TraderPriority: trafficRecord.TraderPriority,
			PlatformFee: trafficRecord.PlatformFee,
			Enabled: trafficRecord.Enabled,
			ActivityParams: &orderpb.TrafficActivityParameters{
				MerchantUnlocked: trafficRecord.ActivityParams.MerchantUnlocked,
				TraderUnlocked: trafficRecord.ActivityParams.TraderUnlocked,
				ManuallyUnlocked: trafficRecord.ActivityParams.ManuallyUnlocked,
				AntifraudUnlocked: trafficRecord.ActivityParams.AntifraudUnlocked,
			},
			AntifraudParams: &orderpb.TrafficAntifraudParameters{
				AntifraudRequired: trafficRecord.AntifraudParams.AntifraudRequired,
			},
			BusinessParams: &orderpb.TrafficBusinessParameters{
				MerchantDealsDuration: durationpb.New(trafficRecord.BusinessParams.MerchantDealsDuration),
			},
		}
	}

	return &orderpb.GetTrafficRecordsResponse{
		TrafficRecords: trafficResponse,
	}, nil
}

func (h *TrafficHandler) DisableTraderTraffic(ctx context.Context, r *orderpb.DisableTraderTrafficRequest) (*orderpb.DisableTraderTrafficResponse, error) {
	traderID := r.TraderId
	if err := h.trafficUsecase.DisableTraderTraffic(traderID); err != nil {
		return nil, err
	}

	return &orderpb.DisableTraderTrafficResponse{}, nil
}

func (h *TrafficHandler) EnableTraderTraffic(ctx context.Context, r *orderpb.EnableTraderTrafficRequest) (*orderpb.EnableTraderTrafficResponse, error) {
	traderID := r.TraderId
	if err := h.trafficUsecase.EnableTraderTraffic(traderID); err != nil {
		return nil, err
	}

	return &orderpb.EnableTraderTrafficResponse{}, nil
}

func (h *TrafficHandler) GetTraderTrafficStatus(ctx context.Context, r *orderpb.GetTraderTrafficStatusRequest) (*orderpb.GetTraderTrafficStatusResponse, error) {
	traderID := r.TraderId
	status, err := h.trafficUsecase.GetTraderTrafficStatus(traderID)
	if err != nil {
		return nil, err
	}

	return &orderpb.GetTraderTrafficStatusResponse{
		Status: status,
	}, nil
}