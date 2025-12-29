package grpcapi

import (
	"context"
	"fmt"
	"log"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TrafficHandler struct {
	orderpb.UnimplementedTrafficServiceServer
	trafficUsecase usecase.TrafficUsecase
}

func NewTrafficHandler(trafficUsecase usecase.TrafficUsecase) *TrafficHandler {
	return &TrafficHandler{trafficUsecase: trafficUsecase}
}

func (h *TrafficHandler) EditTraffic(ctx context.Context, r *orderpb.EditTrafficRequest) (*orderpb.EditTrafficResponse, error) {
    input := &trafficdto.EditTrafficInput{
        ID: r.Id,
    }
    
    // Правильно обрабатываем optional поля
    if r.StoreId != nil {
        input.StoreID = r.StoreId
    }
    
    if r.TraderReward != nil {
        traderReward := float64(*r.TraderReward)
        input.TraderReward = &traderReward
    }
    
    if r.TraderPriority != nil {
        traderPriority := float64(*r.TraderPriority)
        input.TraderPriority = &traderPriority
    }
    
    if r.Enabled != nil {
        input.Enabled = r.Enabled
    }
    
    if r.Name != nil {
        input.Name = r.Name
    }
    
    if r.ActivityParams != nil {
        input.ActivityParams = &trafficdto.TrafficActivityParams{
            TraderUnlocked:    r.ActivityParams.TraderUnlocked,
            AntifraudUnlocked: r.ActivityParams.AntifraudUnlocked,
            ManuallyUnlocked:  r.ActivityParams.ManuallyUnlocked,
        }
    }
    
    if r.AntifraudParams != nil {
        input.AntifraudParams = &trafficdto.TrafficAntifraudParams{
            AntifraudRequired: r.AntifraudParams.AntifraudRequired,
        }
    }
    
    if err := h.trafficUsecase.EditTraffic(input); err != nil {
        return &orderpb.EditTrafficResponse{
            Message: "failed to update traffic",
        }, err
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
			Enabled: trafficRecord.Enabled,
			ActivityParams: &orderpb.TrafficActivityParameters{
				TraderUnlocked: trafficRecord.ActivityParams.TraderUnlocked,
				ManuallyUnlocked: trafficRecord.ActivityParams.ManuallyUnlocked,
				AntifraudUnlocked: trafficRecord.ActivityParams.AntifraudUnlocked,
			},
			AntifraudParams: &orderpb.TrafficAntifraudParameters{
				AntifraudRequired: trafficRecord.AntifraudParams.AntifraudRequired,
			},
			BusinessParams: &orderpb.TrafficBusinessParameters{
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

// GetTraderTraffic получает все записи трафика для трейдера
func (h *TrafficHandler) GetTraderTraffic(ctx context.Context, req *orderpb.GetTraderTrafficRequest) (*orderpb.GetTraderTrafficResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }

    traffics, err := h.trafficUsecase.GetTrafficByTraderID(req.TraderId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get trader traffic: %v", err)
    }

    records := make([]*orderpb.Traffic, 0, len(traffics))
    for _, traffic := range traffics {
        records = append(records, &orderpb.Traffic{
            Id:                  traffic.ID,
            MerchantId:          traffic.MerchantID,
            TraderId:            traffic.TraderID,
            TraderRewardPercent: traffic.TraderRewardPercent,
            TraderPriority:      traffic.TraderPriority,
            Enabled:             traffic.Enabled,
            ActivityParams: &orderpb.TrafficActivityParameters{
                TraderUnlocked:    traffic.ActivityParams.TraderUnlocked,
                AntifraudUnlocked: traffic.ActivityParams.AntifraudUnlocked,
                ManuallyUnlocked:  traffic.ActivityParams.ManuallyUnlocked,
            },
            AntifraudParams: &orderpb.TrafficAntifraudParameters{
                AntifraudRequired: traffic.AntifraudParams.AntifraudRequired,
            },
        })
    }

    return &orderpb.GetTraderTrafficResponse{
        Records: records,
    }, nil
}

// GetTrafficByStore получает трафик по стор ID
func (h *TrafficHandler) GetTrafficByStore(ctx context.Context, req *orderpb.GetTrafficByStoreRequest) (*orderpb.GetTrafficByStoreResponse, error) {
    if req.StoreId == "" {
        return nil, status.Error(codes.InvalidArgument, "store_id is required")
    }
    
    traffics, err := h.trafficUsecase.GetTrafficByStoreID(req.StoreId)
    if err != nil {
        log.Printf("Failed to get traffic by store: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to get traffic: %v", err)
    }
    
    // Проверяем optional поле OnlyActive
    var filteredTraffics []*domain.Traffic
    for _, traffic := range traffics {
        if req.OnlyActive != nil && *req.OnlyActive && !traffic.Enabled {
            continue
        }
        filteredTraffics = append(filteredTraffics, traffic)
    }
    
    trafficResponse := make([]*orderpb.Traffic, len(filteredTraffics))
    for i, traffic := range filteredTraffics {
        trafficResponse[i] = &orderpb.Traffic{
            Id:                  traffic.ID,
            StoreId:            traffic.MerchantStoreID,
            MerchantId:         traffic.MerchantID,
            TraderId:           traffic.TraderID,
            TraderRewardPercent: traffic.TraderRewardPercent,
            TraderPriority:     traffic.TraderPriority,
            Enabled:            traffic.Enabled,
            ActivityParams: &orderpb.TrafficActivityParameters{
                TraderUnlocked:    traffic.ActivityParams.TraderUnlocked,
                ManuallyUnlocked:  traffic.ActivityParams.ManuallyUnlocked,
                AntifraudUnlocked: traffic.ActivityParams.AntifraudUnlocked,
            },
            AntifraudParams: &orderpb.TrafficAntifraudParameters{
                AntifraudRequired: traffic.AntifraudParams.AntifraudRequired,
            },
        }
    }
    
    return &orderpb.GetTrafficByStoreResponse{
        TrafficRecords: trafficResponse,
    }, nil
}

// GetTrafficByMerchant получает трафик по мерчанту
// GetTrafficByMerchant получает трафик по мерчанту
func (h *TrafficHandler) GetTrafficByMerchant(ctx context.Context, req *orderpb.GetTrafficByMerchantRequest) (*orderpb.GetTrafficByMerchantResponse, error) {
    if req.MerchantId == "" {
        return nil, status.Error(codes.InvalidArgument, "merchant_id is required")
    }
    
    traffics, err := h.trafficUsecase.GetTrafficByMerchantID(req.MerchantId)
    if err != nil {
        log.Printf("Failed to get traffic by merchant: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to get traffic: %v", err)
    }
    
    // Проверяем optional поле OnlyActive
    var filteredTraffics []*domain.Traffic
    for _, traffic := range traffics {
        if req.OnlyActive != nil && *req.OnlyActive && !traffic.Enabled {
            continue
        }
        filteredTraffics = append(filteredTraffics, traffic)
    }
    
    trafficResponse := make([]*orderpb.Traffic, len(filteredTraffics))
    for i, traffic := range filteredTraffics {
        trafficResponse[i] = &orderpb.Traffic{
            Id:                  traffic.ID,
            StoreId:            traffic.MerchantStoreID,
            MerchantId:         traffic.MerchantID,
            TraderId:           traffic.TraderID,
            TraderRewardPercent: traffic.TraderRewardPercent,
            TraderPriority:     traffic.TraderPriority,
            Enabled:            traffic.Enabled,
            ActivityParams: &orderpb.TrafficActivityParameters{
                TraderUnlocked:    traffic.ActivityParams.TraderUnlocked,
                ManuallyUnlocked:  traffic.ActivityParams.ManuallyUnlocked,
                AntifraudUnlocked: traffic.ActivityParams.AntifraudUnlocked,
            },
            AntifraudParams: &orderpb.TrafficAntifraudParameters{
                AntifraudRequired: traffic.AntifraudParams.AntifraudRequired,
            },
        }
    }
    
    return &orderpb.GetTrafficByMerchantResponse{
        TrafficRecords: trafficResponse,
    }, nil
}

// GetTrafficByTraderStore получает трафик для конкретного трейдера и стора
func (h *TrafficHandler) GetTrafficByTraderStore(ctx context.Context, req *orderpb.GetTrafficByTraderStoreRequest) (*orderpb.GetTrafficByTraderStoreResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }
    
    if req.StoreId == "" {
        return nil, status.Error(codes.InvalidArgument, "store_id is required")
    }
    
    traffic, err := h.trafficUsecase.GetTrafficByTraderStore(req.TraderId, req.StoreId)
    if err != nil {
        log.Printf("Failed to get traffic by trader and store: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to get traffic: %v", err)
    }
    
    if traffic == nil {
        return &orderpb.GetTrafficByTraderStoreResponse{}, nil
    }
    
    trafficProto := &orderpb.Traffic{
        Id:                  traffic.ID,
        StoreId:            traffic.MerchantStoreID,
        MerchantId:         traffic.MerchantID,
        TraderId:           traffic.TraderID,
        TraderRewardPercent: traffic.TraderRewardPercent,
        TraderPriority:     traffic.TraderPriority,
        Enabled:            traffic.Enabled,
        ActivityParams: &orderpb.TrafficActivityParameters{
            TraderUnlocked:    traffic.ActivityParams.TraderUnlocked,
            ManuallyUnlocked:  traffic.ActivityParams.ManuallyUnlocked,
            AntifraudUnlocked: traffic.ActivityParams.AntifraudUnlocked,
        },
        AntifraudParams: &orderpb.TrafficAntifraudParameters{
            AntifraudRequired: traffic.AntifraudParams.AntifraudRequired,
        },
    }
    
    return &orderpb.GetTrafficByTraderStoreResponse{
        Traffic: trafficProto,
    }, nil
}

// ChangeTrafficStore меняет стор для трафика
func (h *TrafficHandler) ChangeTrafficStore(ctx context.Context, req *orderpb.ChangeTrafficStoreRequest) (*orderpb.ChangeTrafficStoreResponse, error) {
    if req.TrafficId == "" {
        return nil, status.Error(codes.InvalidArgument, "traffic_id is required")
    }
    
    if req.NewStoreId == "" {
        return nil, status.Error(codes.InvalidArgument, "new_store_id is required")
    }
    
    err := h.trafficUsecase.ChangeTrafficStore(req.TrafficId, req.NewStoreId)
    if err != nil {
        log.Printf("Failed to change traffic store: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to change store: %v", err)
    }
    
    return &orderpb.ChangeTrafficStoreResponse{
        Success: true,
    }, nil
}

// Также нужно обновить AddTraffic и EditTraffic для работы с StoreID:

func (h *TrafficHandler) AddTraffic(ctx context.Context, r *orderpb.AddTrafficRequest) (*orderpb.AddTrafficResponse, error) {
    // Теперь используем store_id вместо merchant_id
    traffic := &domain.Traffic{
        MerchantStoreID:	r.StoreId,
        TraderID:           r.TraderId,
        TraderRewardPercent: r.TraderRewardPercent,
        TraderPriority:     r.TraderPriority,
        Enabled:            r.Enabled,
        ActivityParams: domain.TrafficActivityParams{
            TraderUnlocked:    r.ActivityParams.TraderUnlocked,
            AntifraudUnlocked: r.ActivityParams.AntifraudUnlocked,
            ManuallyUnlocked:  r.ActivityParams.ManuallyUnlocked,
        },
        AntifraudParams: domain.TrafficAntifraudParams{
            AntifraudRequired: r.AntifraudParams.AntifraudRequired,
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
