package setup

import (
	"fmt"

	"github.com/LavaJover/shvark-order-service/internal/config"
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
    orderuc "github.com/LavaJover/shvark-order-service/internal/usecase/order"
    disputeuc "github.com/LavaJover/shvark-order-service/internal/usecase/dispute"
)

type UseCases struct {
    OrderUsecase        orderuc.OrderUsecase
    TrafficUsecase      usecase.TrafficUsecase
    BankDetailUsecase   usecase.BankDetailUsecase
    TeamRelationsUsecase usecase.TeamRelationsUsecase
    DeviceUsecase       usecase.DeviceUsecase
    DisputeUsecase      disputeuc.DisputeUsecase
    AutomaticUsecase    usecase.AutomaticUsecase
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
    
    orderUsecase := orderuc.NewDefaultOrderUsecase(
        deps.Repositories.OrderRepo,
        walletHandler,
        trafficUsecase,
        bankDetailUsecase,
        deps.OrderPublisher,
        teamRelationsUsecase,
    )
    
    disputeUsecase := disputeuc.NewDefaultDisputeUsecase(
        deps.Repositories.DisputeRepo,
        walletHandler,
        deps.Repositories.OrderRepo,
        deps.Repositories.TrafficRepo,
        deps.DisputePublisher,
        teamRelationsUsecase,
        deps.Repositories.BankDetailRepo,
    )
    
    automaticUsecase := usecase.NewDefaultAutomaticUsecase(deps.Repositories.OrderRepo)
    
    return &UseCases{
        OrderUsecase:        orderUsecase,
        TrafficUsecase:      trafficUsecase,
        BankDetailUsecase:   bankDetailUsecase,
        TeamRelationsUsecase: teamRelationsUsecase,
        DeviceUsecase:       deviceUsecase,
        DisputeUsecase:      disputeUsecase,
        AutomaticUsecase:    automaticUsecase,
    }, nil
}

func initWalletHandler(cfg *config.OrderConfig) (*handlers.HTTPWalletHandler, error) {
    return handlers.NewHTTPWalletHandler(fmt.Sprintf("%s:%s", cfg.WalletService.Host, cfg.WalletService.Port))
}