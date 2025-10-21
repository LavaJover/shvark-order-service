package grpcapi

import (
	"context"
	"fmt"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
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
		Name: r.Name,
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

	input := &trafficdto.EditTrafficInput{
		ID: r.Id,
		MerchantID: r.MerchantId,
		TraderID: r.TraderId,
		TraderReward: r.TraderReward,
		TraderPriority: r.TraderProirity,
		PlatformFee: r.PlatformFee,
		Enabled: r.Enabled,
		Name: r.Name,
		ActivityParams: &trafficdto.TrafficActivityParams{
			MerchantUnlocked: r.ActivityParams.MerchantUnlocked,
			TraderUnlocked: r.ActivityParams.TraderUnlocked,
			AntifraudUnlocked: r.ActivityParams.AntifraudUnlocked,
			ManuallyUnlocked: r.ActivityParams.ManuallyUnlocked,
		},
		AntifraudParams: &trafficdto.TrafficAntifraudParams{
			AntifraudRequired: r.AntifraudParams.AntifraudRequired,
		},
		BusinessParams: &trafficdto.TrafficBusinessParams{
			MerchantDealsDuration: r.BusinessParams.MerchantDealsDuration.AsDuration(),
		},
	}

	if err := h.trafficUsecase.EditTraffic(input); err != nil {
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
			Name: trafficRecord.Name,
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

func (h *TrafficHandler) SetTraderLockTrafficStatus(ctx context.Context, r *orderpb.SetTraderLockTrafficStatusRequest)(*orderpb.SetTraderLockTrafficStatusResponse, error) {
	err := h.trafficUsecase.SetTraderLockTrafficStatus(r.TraderId, r.Unlocked)
	if err != nil {
		return nil, err
	}

	return &orderpb.SetTraderLockTrafficStatusResponse{}, nil
}

func (h *TrafficHandler) SetMerchantLockTrafficStatus(ctx context.Context, r *orderpb.SetMerchantLockTrafficStatusRequest) (*orderpb.SetMerchantLockTrafficStatusResponse, error) {
	err := h.trafficUsecase.SetMerchantLockTrafficStatus(r.MerchantId, r.Ubnlocked)
	if err != nil {
		return nil, err
	}

	return &orderpb.SetMerchantLockTrafficStatusResponse{}, nil
}

func (h *TrafficHandler) SetManuallyLockTrafficStatus(ctx context.Context, r *orderpb.SetManuallyLockTrafficStatusRequest) (*orderpb.SetManuallyLockTrafficStatusResponse, error) {
	err := h.trafficUsecase.SetManuallyLockTrafficStatus(r.TrafficId, r.Unlocked)
	if err != nil {
		return nil, err
	}

	return &orderpb.SetManuallyLockTrafficStatusResponse{}, nil
}

func (h *TrafficHandler) SetAntifraudLockTrafficStatus(ctx context.Context, r *orderpb.SetAntifraudLockTrafficStatusRequest) (*orderpb.SetAntifraudLockTrafficStatusResponse, error) {
	err := h.trafficUsecase.SetAntifraudLockTrafficStatus(r.TraderId, r.Unlocked)
	if err != nil {
		return nil, err
	}

	return &orderpb.SetAntifraudLockTrafficStatusResponse{}, nil
}

// GetTrafficLockStatuses возвращает все статусы блокировки для указанного трафика
func (h *TrafficHandler) GetTrafficLockStatuses(ctx context.Context, req *orderpb.GetTrafficLockStatusesRequest) (*orderpb.GetTrafficLockStatusesResponse, error) {
	if req.TrafficId == "" {
		return nil, fmt.Errorf("traffic_id is required")
	}

	statuses, err := h.trafficUsecase.GetLockStatuses(req.TrafficId)
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic lock statuses: %w", err)
	}

	return &orderpb.GetTrafficLockStatusesResponse{
		TrafficId:         statuses.TrafficID,
		MerchantUnlocked:  statuses.MerchantUnlocked,
		TraderUnlocked:    statuses.TraderUnlocked,
		AntifraudUnlocked: statuses.AntifraudUnlocked,
		ManuallyUnlocked:  statuses.ManuallyUnlocked,
	}, nil
}

// CheckTrafficUnlocked проверяет, разблокирован ли трафик хотя бы одним способом
func (h *TrafficHandler) CheckTrafficUnlocked(ctx context.Context, req *orderpb.CheckTrafficUnlockedRequest) (*orderpb.CheckTrafficUnlockedResponse, error) {
	if req.TrafficId == "" {
		return nil, fmt.Errorf("traffic_id is required")
	}

	result, err := h.trafficUsecase.IsTrafficUnlocked(req.TrafficId)
	if err != nil {
		return nil, fmt.Errorf("failed to check traffic unlock status: %w", err)
	}

	return &orderpb.CheckTrafficUnlockedResponse{
		TrafficId: result.TrafficID,
		Unlocked:  result.Unlocked,
	}, nil
}
