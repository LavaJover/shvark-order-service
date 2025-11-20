package setup

import (
	"fmt"

	"github.com/LavaJover/shvark-order-service/internal/config"
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
)

type UseCases struct {
    OrderUsecase        usecase.OrderUsecase
    TrafficUsecase      usecase.TrafficUsecase
    BankDetailUsecase   usecase.BankDetailUsecase
    TeamRelationsUsecase usecase.TeamRelationsUsecase
    DeviceUsecase       usecase.DeviceUsecase
    DisputeUsecase      usecase.DisputeUsecase
    AutomaticUsecase    usecase.AutomaticUsecase
    ExchangeRateService    usecase.ExchangeRateService
}

func InitializeUseCases(deps *Dependencies) (*UseCases, error) {
    walletHandler, err := initWalletHandler(deps.Config)
    if err != nil {
        return nil, fmt.Errorf("wallet handler: %w", err)
    }
    
    trafficUsecase := usecase.NewDefaultTrafficUsecase(deps.Repositories.TrafficRepo)
    bankDetailUsecase := usecase.NewDefaultBankDetailUsecase(deps.Repositories.BankDetailRepo)
    teamRelationsUsecase := usecase.NewDefaultTeamRelationsUsecase(deps.Repositories.TeamRelationsRepo)
    deviceUsecase := usecase.NewDefaultDeviceUsecase(deps.Repositories.DeviceRepo)
    
    orderUsecase := usecase.NewDefaultOrderUsecase(
        deps.Repositories.OrderRepo,
        walletHandler,
        trafficUsecase,
        bankDetailUsecase,
        deps.OrderPublisher,
        teamRelationsUsecase,
    )
    
    disputeUsecase := usecase.NewDefaultDisputeUsecase(
        deps.Repositories.DisputeRepo,
        walletHandler,
        deps.Repositories.OrderRepo,
        deps.Repositories.TrafficRepo,
        deps.DisputePublisher,
        teamRelationsUsecase,
        deps.Repositories.BankDetailRepo,
    )
    
    automaticUsecase := usecase.NewDefaultAutomaticUsecase(deps.Repositories.OrderRepo)

    exchangeRateService := usecase.NewDefaultExchangeRateService()
    
    return &UseCases{
        OrderUsecase:        orderUsecase,
        TrafficUsecase:      trafficUsecase,
        BankDetailUsecase:   bankDetailUsecase,
        TeamRelationsUsecase: teamRelationsUsecase,
        DeviceUsecase:       deviceUsecase,
        DisputeUsecase:      disputeUsecase,
        AutomaticUsecase:    automaticUsecase,
        ExchangeRateService: exchangeRateService,
    }, nil
}

func initWalletHandler(cfg *config.OrderConfig) (*handlers.HTTPWalletHandler, error) {
    return handlers.NewHTTPWalletHandler(fmt.Sprintf("%s:%s", cfg.WalletService.Host, cfg.WalletService.Port))
}