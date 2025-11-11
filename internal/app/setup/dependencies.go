package setup

import (
	"fmt"

	"github.com/LavaJover/shvark-order-service/internal/config"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository"
	"gorm.io/gorm"
)

type Dependencies struct {
    Config              *config.OrderConfig
    DB                  *gorm.DB
    OrderPublisher      *publisher.KafkaPublisher
    DisputePublisher    *publisher.KafkaPublisher
    Repositories        *Repositories
}

type Repositories struct {
    OrderRepo         domain.OrderRepository
    BankDetailRepo    domain.BankDetailRepository
    TrafficRepo       domain.TrafficRepository
    TeamRelationsRepo domain.TeamRelationRepository
    DeviceRepo        domain.DeviceRepository
    DisputeRepo       domain.DisputeRepository
    AntiFraudRepo     domain.AntiFraudRepository
}

func InitializeDependencies() (*Dependencies, error) {
    cfg := config.MustLoad()
    
    db := postgres.MustInitDB(cfg)
    
    orderPublisher, err := initOrderPublisher(cfg)
    if err != nil {
        return nil, fmt.Errorf("order publisher: %w", err)
    }
    
    disputePublisher, err := initDisputePublisher(cfg)
    if err != nil {
        return nil, fmt.Errorf("dispute publisher: %w", err)
    }
    
    repos := &Repositories{
        OrderRepo:         repository.NewDefaultOrderRepository(db),
        BankDetailRepo:    repository.NewDefaultBankDetailRepo(db),
        TrafficRepo:       repository.NewDefaultTrafficRepository(db),
        TeamRelationsRepo: repository.NewDefaultTeamRelationsRepository(db),
        DeviceRepo:        repository.NewDefaultDeviceRepository(db),
        DisputeRepo:       repository.NewDefaultDisputeRepository(db),
        AntiFraudRepo:     repository.NewAntiFraudRepository(db),
    }
    
    return &Dependencies{
        Config:           cfg,
        DB:               db,
        OrderPublisher:   orderPublisher,
        DisputePublisher: disputePublisher,
        Repositories:     repos,
    }, nil
}

func initOrderPublisher(cfg *config.OrderConfig) (*publisher.KafkaPublisher, error) {
    config := publisher.KafkaConfig{
        Brokers:   []string{fmt.Sprintf("%s:%s", cfg.KafkaService.Host, cfg.KafkaService.Port)},
        Topic:     "order-events",
        Username:  cfg.KafkaService.Username,
        Password:  cfg.KafkaService.Password,
        Mechanism: cfg.KafkaService.Mechanism,
        TLSEnabled: cfg.KafkaService.TLSEnabled,
    }
    return publisher.NewKafkaPublisher(config)
}

func initDisputePublisher(cfg *config.OrderConfig) (*publisher.KafkaPublisher, error) {
    config := publisher.KafkaConfig{
        Brokers:   []string{fmt.Sprintf("%s:%s", cfg.KafkaService.Host, cfg.KafkaService.Port)},
        Topic:     "dispute-events",
        Username:  cfg.KafkaService.Username,
        Password:  cfg.KafkaService.Password,
        Mechanism: cfg.KafkaService.Mechanism,
        TLSEnabled: cfg.KafkaService.TLSEnabled,
    }
    return publisher.NewKafkaPublisher(config)
}