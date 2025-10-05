package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/config"
	"github.com/LavaJover/shvark-order-service/internal/delivery/grpcapi"
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/engine"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/rules"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/strategies"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/usdt"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("failed to load .env")
	}
	// Reading config
	cfg := config.MustLoad()
	// Init database
	db := postgres.MustInitDB(cfg)

	// Setup kafka
	orderPublisherConfig := publisher.KafkaConfig{
		Brokers:   []string{fmt.Sprintf("%s:%s", cfg.KafkaService.Host, cfg.KafkaService.Port)},
        Topic:     "order-events",
        Username:  cfg.KafkaService.Username,
        Password:  cfg.KafkaService.Password,
        Mechanism: cfg.KafkaService.Mechanism,
    	TLSEnabled: cfg.KafkaService.TLSEnabled,
	}
	orderKafkaPublisher, err := publisher.NewKafkaPublisher(orderPublisherConfig)
	if err != nil {
		log.Fatalf("failed to init kafka order publisher: %v", err)
	}

	disputePublisherConfig := publisher.KafkaConfig{
		Brokers:   []string{fmt.Sprintf("%s:%s", cfg.KafkaService.Host, cfg.KafkaService.Port)},
        Topic:     "dispute-events",
        Username:  cfg.KafkaService.Username,
        Password:  cfg.KafkaService.Password,
        Mechanism: cfg.KafkaService.Mechanism,
    	TLSEnabled: cfg.KafkaService.TLSEnabled,
	}
	disputeKafkaPublisher, err := publisher.NewKafkaPublisher(disputePublisherConfig)
	if err != nil {
		log.Fatalf("failed to init kafka dispute publisher: %v", err)
	}
	// Init order repo
	orderRepo := repository.NewDefaultOrderRepository(db)
	// Init bank detail repo
	bankDetailRepo := repository.NewDefaultBankDetailRepo(db)
	// Init traffic repo
	trafficRepo := repository.NewDefaultTrafficRepository(db)
	// Init team relations repo
	teamRelationsRepo := repository.NewDefaultTeamRelationsRepository(db)
	// Init device repo
	deviceRepo := repository.NewDefaultDeviceRepository(db)

	// Init wallet handler
	httpWalletHandler, err := handlers.NewHTTPWalletHandler(fmt.Sprintf("%s:%s", cfg.WalletService.Host, cfg.WalletService.Port))
	if err != nil {
		log.Fatalf("failed to init wallet usecase")
	}

	// Init traffic usecase
	trafficUsecase := usecase.NewDefaultTrafficUsecase(trafficRepo)
	// Init bank detail usecase
	bankDetailUsecase := usecase.NewDefaultBankDetailUsecase(bankDetailRepo)
	// Init team relations usecase
	teamRelationsUsecase := usecase.NewDefaultTeamRelationsUsecase(teamRelationsRepo)
	// Init order usecase
	uc := usecase.NewDefaultOrderUsecase(orderRepo, httpWalletHandler, trafficUsecase, bankDetailUsecase, orderKafkaPublisher, teamRelationsUsecase)
	// Init device usecase
	deviceUsecase := usecase.NewDefaultDeviceUsecase(deviceRepo)

	// dispute
	disputeRepo := repository.NewDefaultDisputeRepository(db)
	disputeUc := usecase.NewDefaultDisputeUsecase(
		disputeRepo,
		httpWalletHandler,
		orderRepo,
		trafficRepo,
		disputeKafkaPublisher,
		teamRelationsUsecase,
		bankDetailRepo,
	)

	// Creating gRPC server
	grpcServer := grpc.NewServer()
	orderHandler := grpcapi.NewOrderHandler(uc, disputeUc, bankDetailUsecase)
	trafficHandler := grpcapi.NewTrafficHandler(trafficUsecase)
	bankDetailHandler := grpcapi.NewBankDetailHandler(bankDetailUsecase)
	teamRelationsHandler := grpcapi.NewTeamRelationsHandler(teamRelationsUsecase)
	deviceHandler := grpcapi.NewDeviceHandler(deviceUsecase)

	orderpb.RegisterOrderServiceServer(grpcServer, orderHandler)
	orderpb.RegisterTrafficServiceServer(grpcServer, trafficHandler)
	orderpb.RegisterBankDetailServiceServer(grpcServer, bankDetailHandler)
	orderpb.RegisterTeamRelationsServiceServer(grpcServer, teamRelationsHandler)
	orderpb.RegisterDeviceServiceServer(grpcServer, deviceHandler)

	// Start
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.GRPCServer.Host, cfg.GRPCServer.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// order auto-cancel
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			<-ticker.C
			err := uc.CancelExpiredOrders(context.Background())
			if err != nil {
				log.Printf("Auto-cancel error: %v\n", err)
			}
		}
	}()

	// updating crypto-rates
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for {
			usdtRate, err := usdt.GET_USDT_RUB_RATES(5)
			if err != nil {
				slog.Error("USD/RUB rates update failed", "error", err.Error())
				continue
			}
			slog.Info("USD/RUB rates updated", "usdt/rub", usdtRate)
			<-ticker.C
		}
	}()

	// auto accept expired disputes
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for {
			<-ticker.C
			err := disputeUc.AcceptExpiredDisputes()
			if err != nil {
				log.Printf("Auto-accept dispute error: %v\n", err)
			}
		}
	}()

	// Init antifraud
	antifraudLogger := slog.Default()
	antifraudEngine := engine.NewAntiFraudEngine(db, antifraudLogger)

	antifraudEngine.RegisterStrategy(strategies.NewConsecutiveOrdersStrategy(db))
	antifraudEngine.RegisterStrategy(strategies.NewCanceledOrdersStrategy(db))

	ruleManager := engine.NewRuleManager(db)

	  // Создаем правила
	  consecutiveConfig := &rules.ConsecutiveOrdersConfig{
        MaxConsecutiveOrders: 10,
        TimeWindow:          24 * time.Hour,
        StatesToCount:       []string{"CANCELED"},
    }

    ruleManager.CreateRule(context.Background(), 
        "Max Consecutive Orders", 
        "consecutive_orders", 
        consecutiveConfig, 
        100)

    canceledConfig := &rules.CanceledOrdersConfig{
        MaxCanceledOrders: 5,
        TimeWindow:        24 * time.Hour,
        CanceledStatuses:  []string{"CANCELED"},
    }

    ruleManager.CreateRule(context.Background(), 
        "Max Canceled Orders", 
        "canceled_orders", 
        canceledConfig, 
		90)

	// Запускаем планировщик для автоматических проверок
	scheduler := engine.NewScheduler(antifraudEngine, db, 10*time.Second, antifraudLogger)
	go scheduler.Start(context.Background())

	log.Printf("gRPC server started on %s:%s\n", cfg.GRPCServer.Host, cfg.GRPCServer.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
