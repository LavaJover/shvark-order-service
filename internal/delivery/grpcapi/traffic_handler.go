package grpcapi

import (
	"context"
	"fmt"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

type TrafficHandler struct {
	orderpb.UnimplementedTrafficServiceServer
	trafficUsecase usecase.TrafficUsecase
}

func NewTrafficHandler(trafficUsecase usecase.TrafficUsecase) *TrafficHandler {
	return &TrafficHandler{trafficUsecase: trafficUsecase}
}

// Обновляем AddTraffic для поддержки конфигурации курсов
// Обновляем AddTraffic для поддержки конфигурации курсов
func (h *TrafficHandler) AddTraffic(ctx context.Context, r *orderpb.AddTrafficRequest) (*orderpb.AddTrafficResponse, error) {
    traffic := &domain.Traffic{
        MerchantID:          r.MerchantId,
        TraderID:            r.TraderId,
        TraderRewardPercent: r.TraderRewardPercent,
        TraderPriority:      r.TraderPriority,
        Enabled:             r.Enabled,
        PlatformFee:         r.PlatformFee,
        Name:                r.Name,
        ActivityParams: domain.TrafficActivityParams{
            MerchantUnlocked:  r.ActivityParams.MerchantUnlocked,
            TraderUnlocked:    r.ActivityParams.TraderUnlocked,
            AntifraudUnlocked: r.ActivityParams.AntifraudUnlocked,
            ManuallyUnlocked:  r.ActivityParams.ManuallyUnlocked,
        },
        AntifraudParams: domain.TrafficAntifraudParams{
            AntifraudRequired: r.AntifraudParams.AntifraudRequired,
        },
        BusinessParams: domain.TrafficBusinessParams{
            MerchantDealsDuration: r.BusinessParams.MerchantDealsDuration.AsDuration(),
        },
    }

    // НОВОЕ: Добавляем конфигурацию курсов если передана
    if r.ExchangeConfig != nil {
        traffic.ExchangeConfig = &domain.ExchangeConfig{
            ExchangeProvider:  r.ExchangeConfig.ExchangeProvider,
            MarkupPercent:     r.ExchangeConfig.MarkupPercent,
            CurrencyPair:      r.ExchangeConfig.CurrencyPair,
            FallbackProviders: r.ExchangeConfig.FallbackProviders,
        }

        if r.ExchangeConfig.OrderBookRange != nil {
            traffic.ExchangeConfig.OrderBookPositions = &domain.OrderBookRange{
                Start: int(r.ExchangeConfig.OrderBookRange.Start),
                End:   int(r.ExchangeConfig.OrderBookRange.End),
            }
        }
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

// Обновляем EditTraffic для поддержки конфигурации курсов
// Обновляем EditTraffic для поддержки конфигурации курсов
func (h *TrafficHandler) EditTraffic(ctx context.Context, r *orderpb.EditTrafficRequest) (*orderpb.EditTrafficResponse, error) {
    input := &trafficdto.EditTrafficInput{
        ID:             r.Id,
        MerchantID:     r.MerchantId,
        TraderID:       r.TraderId,
        TraderReward:   r.TraderReward,
        TraderPriority: r.TraderProirity,
        PlatformFee:    r.PlatformFee,
        Enabled:        r.Enabled,
        Name:           r.Name,
        ActivityParams: &trafficdto.TrafficActivityParams{},
        AntifraudParams: &trafficdto.TrafficAntifraudParams{},
        BusinessParams: &trafficdto.TrafficBusinessParams{},
    }

    if r.ActivityParams != nil {
        input.ActivityParams.MerchantUnlocked = r.ActivityParams.MerchantUnlocked
        input.ActivityParams.TraderUnlocked = r.ActivityParams.TraderUnlocked
        input.ActivityParams.AntifraudUnlocked = r.ActivityParams.AntifraudUnlocked
        input.ActivityParams.ManuallyUnlocked = r.ActivityParams.ManuallyUnlocked
    }
    if r.AntifraudParams != nil {
        input.AntifraudParams.AntifraudRequired = r.AntifraudParams.AntifraudRequired
    }
    if r.BusinessParams != nil {
        input.BusinessParams.MerchantDealsDuration = r.BusinessParams.MerchantDealsDuration.AsDuration()
    }

    // НОВОЕ: Добавляем конфигурацию курсов если передана
    if r.ExchangeConfig != nil {
        input.ExchangeConfig = &trafficdto.ExchangeConfigInput{
            ExchangeProvider:  r.ExchangeConfig.ExchangeProvider,
            MarkupPercent:     r.ExchangeConfig.MarkupPercent,
            CurrencyPair:      r.ExchangeConfig.CurrencyPair,
            FallbackProviders: r.ExchangeConfig.FallbackProviders,
        }

        if r.ExchangeConfig.OrderBookRange != nil {
            input.ExchangeConfig.OrderBookRange = &trafficdto.OrderBookRangeInput{
                Start: int(r.ExchangeConfig.OrderBookRange.Start),
                End:   int(r.ExchangeConfig.OrderBookRange.End),
            }
        }
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

// func (h *TrafficHandler) GetTrafficRecords(ctx context.Context, r *orderpb.GetTrafficRecordsRequest) (*orderpb.GetTrafficRecordsResponse, error) {
// 	page, limit := r.Page, r.Limit
// 	trafficRecords, err := h.trafficUsecase.GetTrafficRecords(page, limit)
// 	if err != nil {
// 		return nil, err
// 	}

// 	trafficResponse := make([]*orderpb.Traffic, len(trafficRecords))
// 	for i, trafficRecord := range trafficRecords {
// 		trafficResponse[i] = &orderpb.Traffic{
// 			Id: trafficRecord.ID,
// 			MerchantId: trafficRecord.MerchantID,
// 			TraderId: trafficRecord.TraderID,
// 			TraderRewardPercent: trafficRecord.TraderRewardPercent,
// 			TraderPriority: trafficRecord.TraderPriority,
// 			PlatformFee: trafficRecord.PlatformFee,
// 			Enabled: trafficRecord.Enabled,
// 			Name: trafficRecord.Name,
// 			ActivityParams: &orderpb.TrafficActivityParameters{
// 				MerchantUnlocked: trafficRecord.ActivityParams.MerchantUnlocked,
// 				TraderUnlocked: trafficRecord.ActivityParams.TraderUnlocked,
// 				ManuallyUnlocked: trafficRecord.ActivityParams.ManuallyUnlocked,
// 				AntifraudUnlocked: trafficRecord.ActivityParams.AntifraudUnlocked,
// 			},
// 			AntifraudParams: &orderpb.TrafficAntifraudParameters{
// 				AntifraudRequired: trafficRecord.AntifraudParams.AntifraudRequired,
// 			},
// 			BusinessParams: &orderpb.TrafficBusinessParameters{
// 				MerchantDealsDuration: durationpb.New(trafficRecord.BusinessParams.MerchantDealsDuration),
// 			},
// 		}
// 	}

// 	return &orderpb.GetTrafficRecordsResponse{
// 		TrafficRecords: trafficResponse,
// 	}, nil
// }

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
            PlatformFee:         traffic.PlatformFee,
            TraderPriority:      traffic.TraderPriority,
            Enabled:             traffic.Enabled,
            Name:                traffic.Name,
            ActivityParams: &orderpb.TrafficActivityParameters{
                MerchantUnlocked:  traffic.ActivityParams.MerchantUnlocked,
                TraderUnlocked:    traffic.ActivityParams.TraderUnlocked,
                AntifraudUnlocked: traffic.ActivityParams.AntifraudUnlocked,
                ManuallyUnlocked:  traffic.ActivityParams.ManuallyUnlocked,
            },
            AntifraudParams: &orderpb.TrafficAntifraudParameters{
                AntifraudRequired: traffic.AntifraudParams.AntifraudRequired,
            },
            BusinessParams: &orderpb.TrafficBusinessParameters{
                MerchantDealsDuration: durationpb.New(traffic.BusinessParams.MerchantDealsDuration),
            },
        })
    }

    return &orderpb.GetTraderTrafficResponse{
        Records: records,
    }, nil
}

// НОВЫЕ МЕТОДЫ ДЛЯ УПРАВЛЕНИЯ КУРСАМИ

// UpdateExchangeConfig обновляет конфигурацию курсов для трафика
// UpdateExchangeConfig обновляет конфигурацию курсов для трафика
func (h *TrafficHandler) UpdateExchangeConfig(ctx context.Context, req *orderpb.UpdateExchangeConfigRequest) (*orderpb.UpdateExchangeConfigResponse, error) {
    if req.TrafficId == "" {
        return nil, status.Error(codes.InvalidArgument, "traffic_id is required")
    }

    if req.ExchangeConfig == nil {
        return nil, status.Error(codes.InvalidArgument, "exchange_config is required")
    }

    config := &domain.ExchangeConfig{
        ExchangeProvider:  req.ExchangeConfig.ExchangeProvider,
        MarkupPercent:     req.ExchangeConfig.MarkupPercent,
        CurrencyPair:      req.ExchangeConfig.CurrencyPair,
        FallbackProviders: req.ExchangeConfig.FallbackProviders,
    }

    if req.ExchangeConfig.OrderBookRange != nil {
        config.OrderBookPositions = &domain.OrderBookRange{
            Start: int(req.ExchangeConfig.OrderBookRange.Start),
            End:   int(req.ExchangeConfig.OrderBookRange.End),
        }
    }

    if err := h.trafficUsecase.UpdateExchangeConfig(req.TrafficId, config); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to update exchange config: %v", err)
    }

    return &orderpb.UpdateExchangeConfigResponse{
        Success: true,
        Message: "Exchange configuration updated successfully",
    }, nil
}

// GetExchangeConfig получает конфигурацию курсов для трафика
// GetExchangeConfig получает конфигурацию курсов для трафика
func (h *TrafficHandler) GetExchangeConfig(ctx context.Context, req *orderpb.GetExchangeConfigRequest) (*orderpb.GetExchangeConfigResponse, error) {
    if req.TrafficId == "" {
        return nil, status.Error(codes.InvalidArgument, "traffic_id is required")
    }

    config, err := h.trafficUsecase.GetExchangeConfig(req.TrafficId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get exchange config: %v", err)
    }

    // ИСПРАВЛЕНИЕ: Создаем ExchangeConfig как вложенную структуру
    exchangeConfig := &orderpb.ExchangeConfig{
        ExchangeProvider:  config.ExchangeProvider,
        MarkupPercent:     config.MarkupPercent,
        CurrencyPair:      config.CurrencyPair,
        FallbackProviders: config.FallbackProviders,
    }

    if config.OrderBookPositions != nil {
        exchangeConfig.OrderBookRange = &orderpb.OrderBookRange{
            Start: int32(config.OrderBookPositions.Start),
            End:   int32(config.OrderBookPositions.End),
        }
    }

    response := &orderpb.GetExchangeConfigResponse{
        ExchangeConfig: exchangeConfig, // Правильное присвоение
    }

    return response, nil
}

// GetAvailableExchangeProviders возвращает список доступных провайдеров
func (h *TrafficHandler) GetAvailableExchangeProviders(ctx context.Context, req *orderpb.GetAvailableExchangeProvidersRequest) (*orderpb.GetAvailableExchangeProvidersResponse, error) {
    providers := h.trafficUsecase.GetAvailableExchangeProviders()
    
    return &orderpb.GetAvailableExchangeProvidersResponse{
        Providers: providers,
    }, nil
}

// Обновляем GetTrafficRecords для включения конфигурации курсов в ответ
// Обновляем GetTrafficRecords для включения конфигурации курсов в ответ
func (h *TrafficHandler) GetTrafficRecords(ctx context.Context, r *orderpb.GetTrafficRecordsRequest) (*orderpb.GetTrafficRecordsResponse, error) {
    page, limit := r.Page, r.Limit
    trafficRecords, err := h.trafficUsecase.GetTrafficRecords(page, limit)
    if err != nil {
        return nil, err
    }

    trafficResponse := make([]*orderpb.Traffic, len(trafficRecords))
    for i, trafficRecord := range trafficRecords {
        trafficPB := &orderpb.Traffic{
            Id:                  trafficRecord.ID,
            MerchantId:          trafficRecord.MerchantID,
            TraderId:            trafficRecord.TraderID,
            TraderRewardPercent: trafficRecord.TraderRewardPercent,
            TraderPriority:      trafficRecord.TraderPriority,
            PlatformFee:         trafficRecord.PlatformFee,
            Enabled:             trafficRecord.Enabled,
            Name:                trafficRecord.Name,
            ActivityParams: &orderpb.TrafficActivityParameters{
                MerchantUnlocked:  trafficRecord.ActivityParams.MerchantUnlocked,
                TraderUnlocked:    trafficRecord.ActivityParams.TraderUnlocked,
                ManuallyUnlocked:  trafficRecord.ActivityParams.ManuallyUnlocked,
                AntifraudUnlocked: trafficRecord.ActivityParams.AntifraudUnlocked,
            },
            AntifraudParams: &orderpb.TrafficAntifraudParameters{
                AntifraudRequired: trafficRecord.AntifraudParams.AntifraudRequired,
            },
            BusinessParams: &orderpb.TrafficBusinessParameters{
                MerchantDealsDuration: durationpb.New(trafficRecord.BusinessParams.MerchantDealsDuration),
            },
        }

        // НОВОЕ: Добавляем конфигурацию курсов в ответ
        if trafficRecord.ExchangeConfig != nil {
            trafficPB.ExchangeConfig = &orderpb.ExchangeConfig{
                ExchangeProvider:  trafficRecord.ExchangeConfig.ExchangeProvider,
                MarkupPercent:     trafficRecord.ExchangeConfig.MarkupPercent,
                CurrencyPair:      trafficRecord.ExchangeConfig.CurrencyPair,
                FallbackProviders: trafficRecord.ExchangeConfig.FallbackProviders,
            }

            if trafficRecord.ExchangeConfig.OrderBookPositions != nil {
                trafficPB.ExchangeConfig.OrderBookRange = &orderpb.OrderBookRange{
                    Start: int32(trafficRecord.ExchangeConfig.OrderBookPositions.Start),
                    End:   int32(trafficRecord.ExchangeConfig.OrderBookPositions.End),
                }
            }
        }

        trafficResponse[i] = trafficPB
    }

    return &orderpb.GetTrafficRecordsResponse{
        TrafficRecords: trafficResponse,
    }, nil
}