package background

import (
    "context"
    "log"
    "time"
    
    "github.com/LavaJover/shvark-order-service/internal/infrastructure/usdt"
    "github.com/LavaJover/shvark-order-service/internal/usecase"
)

type BackgroundTasks struct {
    OrderUsecase    usecase.OrderUsecase
    DisputeUsecase  usecase.DisputeUsecase
    DeviceUsecase   usecase.DeviceUsecase
}

func NewBackgroundTasks(orderUC usecase.OrderUsecase, disputeUC usecase.DisputeUsecase, deviceUC usecase.DeviceUsecase) *BackgroundTasks {
    return &BackgroundTasks{
        OrderUsecase:   orderUC,
        DisputeUsecase: disputeUC,
        DeviceUsecase:  deviceUC,
    }
}

func (bt *BackgroundTasks) StartAll(ctx context.Context) {
    go bt.startOrderAutoCancel(ctx)
    go bt.startCryptoRatesUpdate(ctx)
    go bt.startAutoAcceptExpiredDisputes(ctx)
    go bt.startDeviceOfflineCheck(ctx)
}

func (bt *BackgroundTasks) startOrderAutoCancel(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := bt.OrderUsecase.CancelExpiredOrders(ctx); err != nil {
                log.Printf("Auto-cancel error: %v\n", err)
            }
        }
    }
}

func (bt *BackgroundTasks) startCryptoRatesUpdate(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            usdtRate, err := usdt.GET_USDT_RUB_RATES(5)
            if err != nil {
                log.Printf("USD/RUB rates update failed: %v", err)
                continue
            }
            log.Printf("USD/RUB rates updated: usdt/rub=%.2f", usdtRate)
        }
    }
}

func (bt *BackgroundTasks) startAutoAcceptExpiredDisputes(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := bt.DisputeUsecase.AcceptExpiredDisputes(); err != nil {
                log.Printf("Auto-accept dispute error: %v\n", err)
            }
        }
    }
}

func (bt *BackgroundTasks) startDeviceOfflineCheck(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := bt.DeviceUsecase.CheckOfflineDevices(); err != nil {
                log.Printf("Error checking offline devices: %v", err)
            }
        }
    }
}