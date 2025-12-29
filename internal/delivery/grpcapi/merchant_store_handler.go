package grpcapi

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	merchantdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/merchant"
	trafficdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/traffic"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen/order"
)

type MerchantStoreHandler struct {
    orderpb.UnimplementedMerchantStoreServiceServer
    merchantStoreUC usecase.MerchantStoreUsecase
    trafficUC       usecase.TrafficUsecase
}

func NewMerchantStoreHandler(
    merchantStoreUC usecase.MerchantStoreUsecase,
    trafficUC usecase.TrafficUsecase,
) *MerchantStoreHandler {
    return &MerchantStoreHandler{
        merchantStoreUC: merchantStoreUC,
        trafficUC:       trafficUC,
    }
}

// CreateStore создает новый стор для мерчанта
func (h *MerchantStoreHandler) CreateStore(ctx context.Context, req *orderpb.CreateStoreRequest) (*orderpb.CreateStoreResponse, error) {
    if req.MerchantId == "" {
        return nil, status.Error(codes.InvalidArgument, "merchant_id is required")
    }
    
    if req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "store name is required")
    }
    
    input := &merchantdto.CreateStoreInput{
        MerchantID:    req.MerchantId,
        Name:          req.Name,
        PlatformFee:   float64(req.PlatformFee),
        DealsDuration: req.DealsDuration.AsDuration(),
        Description:   req.Description,
        Category:      req.Category,
        MaxDailyDeals: int(req.MaxDailyDeals),
        MinDealAmount: float64(req.MinDealAmount),
        MaxDealAmount: float64(req.MaxDealAmount),
        Currency:      req.Currency,
    }
    
    store, err := h.merchantStoreUC.CreateMerchantStore(input)
    if err != nil {
        log.Printf("Failed to create merchant store: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to create store: %v", err)
    }
    
    return &orderpb.CreateStoreResponse{
        Store: h.storeToProto(store),
    }, nil
}

// UpdateStore обновляет стор
func (h *MerchantStoreHandler) UpdateStore(ctx context.Context, req *orderpb.UpdateStoreRequest) (*orderpb.UpdateStoreResponse, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "store id is required")
    }
    
    input := &merchantdto.UpdateStoreInput{
        ID: req.Id,
    }
    
    // Проверяем optional поля
    if req.Name != nil {
        input.Name = *req.Name
    }
    
    if req.PlatformFee != nil {
        input.PlatformFee = float64(*req.PlatformFee)
    }
    
    if req.IsActive != nil {
        input.IsActive = *req.IsActive
    }
    
    if req.DealsDuration != nil {
        input.DealsDuration = req.DealsDuration.AsDuration()
    }
    
    if req.Description != nil {
        input.Description = *req.Description
    }
    
    if req.Category != nil {
        input.Category = *req.Category
    }
    
    if req.MaxDailyDeals != nil {
        input.MaxDailyDeals = int(*req.MaxDailyDeals)
    }
    
    if req.MinDealAmount != nil {
        input.MinDealAmount = float64(*req.MinDealAmount)
    }
    
    if req.MaxDealAmount != nil {
        input.MaxDealAmount = float64(*req.MaxDealAmount)
    }
    
    if req.Currency != nil {
        input.Currency = *req.Currency
    }
    
    store, err := h.merchantStoreUC.UpdateMerchantStore(input)
    if err != nil {
        log.Printf("Failed to update merchant store: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to update store: %v", err)
    }
    
    return &orderpb.UpdateStoreResponse{
        Store: h.storeToProto(store),
    }, nil
}

// DeleteStore удаляет стор
func (h *MerchantStoreHandler) DeleteStore(ctx context.Context, req *orderpb.DeleteStoreRequest) (*orderpb.DeleteStoreResponse, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "store id is required")
    }
    
    err := h.merchantStoreUC.DeleteMerchantStore(req.Id)
    if err != nil {
        log.Printf("Failed to delete merchant store: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to delete store: %v", err)
    }
    
    return &orderpb.DeleteStoreResponse{
        Success: true,
    }, nil
}

// GetStore получает стор по ID
func (h *MerchantStoreHandler) GetStore(ctx context.Context, req *orderpb.GetStoreRequest) (*orderpb.GetStoreResponse, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "store id is required")
    }
    
    store, err := h.merchantStoreUC.GetMerchantStoreByID(req.Id)
    if err != nil {
        log.Printf("Failed to get merchant store: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to get store: %v", err)
    }
    
    if store == nil {
        return nil, status.Error(codes.NotFound, "store not found")
    }
    
    return &orderpb.GetStoreResponse{
        Store: h.storeToProto(store),
    }, nil
}

// ListStores получает список сторов с пагинацией
func (h *MerchantStoreHandler) ListStores(ctx context.Context, req *orderpb.ListStoresRequest) (*orderpb.ListStoresResponse, error) {
    if req.Page < 1 {
        req.Page = 1
    }
    
    if req.Limit < 1 || req.Limit > 100 {
        req.Limit = 20
    }
    
    stores, err := h.merchantStoreUC.GetMerchantStores(req.Page, req.Limit)
    if err != nil {
        log.Printf("Failed to list merchant stores: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to list stores: %v", err)
    }
    
    var protoStores []*orderpb.Store
    for _, store := range stores {
        // Проверяем optional поле category
        if req.Category != nil && *req.Category != "" && store.Category != *req.Category {
            continue
        }
        
        // Проверяем optional поле only_active
        if req.OnlyActive != nil && *req.OnlyActive && !store.IsActive {
            continue
        }
        
        protoStores = append(protoStores, h.storeToProto(store))
    }
    
    total := int32(len(protoStores))
    
    return &orderpb.ListStoresResponse{
        Stores: protoStores,
        Total:  total,
        Page:   req.Page,
        Limit:  req.Limit,
    }, nil
}

// GetStoresByMerchant получает сторы по мерчанту
func (h *MerchantStoreHandler) GetStoresByMerchant(ctx context.Context, req *orderpb.GetStoresByMerchantRequest) (*orderpb.GetStoresByMerchantResponse, error) {
    if req.MerchantId == "" {
        return nil, status.Error(codes.InvalidArgument, "merchant_id is required")
    }

    var stores []*domain.MerchantStore
    var err error

    // only_active - обычное bool (не optional)
    if req.OnlyActive != nil && *req.OnlyActive{
        stores, err = h.merchantStoreUC.GetActiveStoresByMerchantID(req.MerchantId)
    } else {
        stores, err = h.merchantStoreUC.GetStoresByMerchantID(req.MerchantId)
    }

    if err != nil {
        log.Printf("Failed to get merchant stores: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to get merchant stores: %v", err)
    }

    var protoStores []*orderpb.Store
    for _, store := range stores {
        protoStores = append(protoStores, h.storeToProto(store))
    }

    return &orderpb.GetStoresByMerchantResponse{
        Stores: protoStores,
    }, nil
}

// ConnectTrader подключает трейдера к стору
func (h *MerchantStoreHandler) ConnectTrader(ctx context.Context, req *orderpb.ConnectTraderRequest) (*orderpb.ConnectTraderResponse, error) {
    if req.StoreId == "" {
        return nil, status.Error(codes.InvalidArgument, "store_id is required")
    }
    
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }
    
    // Получаем стор для получения параметров
    store, err := h.merchantStoreUC.GetMerchantStoreByID(req.StoreId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get store: %v", err)
    }
    
    if store == nil {
        return nil, status.Error(codes.NotFound, "store not found")
    }
    
    // Проверяем, нет ли уже подключения
    existingTraffic, err := h.trafficUC.GetTrafficByTraderStore(req.TraderId, req.StoreId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to check existing traffic: %v", err)
    }
    
    if existingTraffic != nil {
        return nil, status.Error(codes.AlreadyExists, "trader is already connected to this store")
    }
    
    // Создаем трафик
    traffic := &domain.Traffic{
        ID:                  uuid.New().String(),
        MerchantStoreID:             store.ID,
        MerchantID:          store.MerchantID,
        TraderID:            req.TraderId,
        TraderRewardPercent: float64(req.TraderRewardPercent),
        TraderPriority:      1.0,
        Enabled:             req.Enabled,
        ActivityParams: domain.TrafficActivityParams{
            TraderUnlocked: false,
            AntifraudUnlocked: false,
            ManuallyUnlocked: false,
        },
        CreatedAt:           time.Now(),
        UpdatedAt:           time.Now(),
    }
    
    if err := h.trafficUC.AddTraffic(traffic); err != nil {
        log.Printf("Failed to connect trader to store: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to connect trader: %v", err)
    }
    
    return &orderpb.ConnectTraderResponse{
        Success:   true,
        TrafficId: traffic.ID,
    }, nil
}

// UpdateTraderConnection обновляет параметры подключения трейдера
func (h *MerchantStoreHandler) UpdateTraderConnection(ctx context.Context, req *orderpb.UpdateTraderConnectionRequest) (*orderpb.UpdateTraderConnectionResponse, error) {
    if req.TrafficId == "" {
        return nil, status.Error(codes.InvalidArgument, "traffic_id is required")
    }
    
    // Получаем существующий трафик
    traffic, err := h.trafficUC.GetTrafficByID(req.TrafficId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get traffic: %v", err)
    }
    
    if traffic == nil {
        return nil, status.Error(codes.NotFound, "traffic not found")
    }
    
    // Создаем DTO для обновления
    input := &trafficdto.EditTrafficInput{
        ID: req.TrafficId,
    }
    
    // Проверяем optional поля
    if req.TraderRewardPercent != nil {
        traderReward := float64(*req.TraderRewardPercent)
        input.TraderReward = &traderReward
    }
    
    if req.Enabled != nil {
        input.Enabled = req.Enabled
    }
    
    // merchant_unlocked отсутствует в proto, но trader_unlocked есть
    if req.TraderUnlocked != nil {
        traderUnlocked := *req.TraderUnlocked
        input.ActivityParams.TraderUnlocked = traderUnlocked
    }
    
    if req.AntifraudUnlocked != nil {
        antifraudUnlocked := *req.AntifraudUnlocked
        input.ActivityParams.AntifraudUnlocked = antifraudUnlocked
    }
    
    if req.ManuallyUnlocked != nil {
        manuallyUnlocked := *req.ManuallyUnlocked
        input.ActivityParams.ManuallyUnlocked = manuallyUnlocked
    }
    
    // Обновляем трафик
    if err := h.trafficUC.EditTraffic(input); err != nil {
        log.Printf("Failed to update trader connection: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to update connection: %v", err)
    }
    
    return &orderpb.UpdateTraderConnectionResponse{
        Success: true,
    }, nil
}

// DisconnectTrader отключает трейдера от стора
func (h *MerchantStoreHandler) DisconnectTrader(ctx context.Context, req *orderpb.DisconnectTraderRequest) (*orderpb.DisconnectTraderResponse, error) {
    if req.TrafficId == "" {
        return nil, status.Error(codes.InvalidArgument, "traffic_id is required")
    }
    
    if err := h.trafficUC.DeleteTraffic(req.TrafficId); err != nil {
        log.Printf("Failed to disconnect trader: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to disconnect trader: %v", err)
    }
    
    return &orderpb.DisconnectTraderResponse{
        Success: true,
    }, nil
}

// GetStoreTraffic получает список трейдеров, подключенных к стору
func (h *MerchantStoreHandler) GetStoreTraffic(ctx context.Context, req *orderpb.GetStoreTrafficRequest) (*orderpb.GetStoreTrafficResponse, error) {
    if req.StoreId == "" {
        return nil, status.Error(codes.InvalidArgument, "store_id is required")
    }
    
    traffics, err := h.trafficUC.GetTrafficByStoreID(req.StoreId)
    if err != nil {
        log.Printf("Failed to get store traffic: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to get store traffic: %v", err)
    }
    
    var trafficInfos []*orderpb.TrafficInfo
    for _, traffic := range traffics {
        // only_active - обычное bool (не optional)
        if req.OnlyActive && !traffic.Enabled {
            continue
        }
        
        trafficInfos = append(trafficInfos, &orderpb.TrafficInfo{
            TrafficId:           traffic.ID,
            StoreId:            traffic.MerchantStoreID,
            MerchantId:         traffic.MerchantID,
            TraderId:           traffic.TraderID,
            TraderRewardPercent: float32(traffic.TraderRewardPercent),
            Enabled:            traffic.Enabled,
            CreatedAt:          timestamppb.New(traffic.CreatedAt),
            UpdatedAt:          timestamppb.New(traffic.UpdatedAt),
        })
    }
    
    return &orderpb.GetStoreTrafficResponse{
        Traffics: trafficInfos,
    }, nil
}

// GetTraderStores получает сторы, к которым подключен трейдер
func (h *MerchantStoreHandler) GetTraderStores(ctx context.Context, req *orderpb.GetTraderStoresRequest) (*orderpb.GetTraderStoresResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }
    
    // Получаем все трафики трейдера
    traffics, err := h.trafficUC.GetTrafficByTraderID(req.TraderId)
    if err != nil {
        log.Printf("Failed to get trader traffic: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to get trader stores: %v", err)
    }
    
    // Собираем уникальные store IDs
    storeIDs := make(map[string]bool)
    for _, traffic := range traffics {
        // Проверяем optional поле only_active
        if req.OnlyActive != nil && *req.OnlyActive && !traffic.Enabled {
            continue
        }
        storeIDs[traffic.MerchantStoreID] = true
    }
    
    // Получаем сторы по IDs
    var stores []*orderpb.Store
    for storeID := range storeIDs {
        store, err := h.merchantStoreUC.GetMerchantStoreByID(storeID)
        if err != nil {
            log.Printf("Failed to get store %s: %v", storeID, err)
            continue
        }
        
        if store != nil {
            stores = append(stores, h.storeToProto(store))
        }
    }
    
    return &orderpb.GetTraderStoresResponse{
        Stores: stores,
    }, nil
}

// Вспомогательные методы

func (h *MerchantStoreHandler) storeToProto(store *domain.MerchantStore) *orderpb.Store {
    if store == nil {
        return nil
    }
    
    return &orderpb.Store{
        Id:              store.ID,
        MerchantId:      store.MerchantID,
        Name:            store.Name,
        PlatformFee:     float32(store.PlatformFee),
        IsActive:        store.IsActive,
        DealsDuration:   durationpb.New(store.DealsDuration),
        Description:     store.Description,
        Category:        store.Category,
        MaxDailyDeals:   int32(store.MaxDailyDeals),
        MinDealAmount:   float32(store.MinDealAmount),
        MaxDealAmount:   float32(store.MaxDealAmount),
        Currency:        store.Currency,
        CreatedAt:       timestamppb.New(store.CreatedAt),
        UpdatedAt:       timestamppb.New(store.UpdatedAt),
    }
}