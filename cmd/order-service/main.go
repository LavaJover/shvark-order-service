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

	brokers := []string{fmt.Sprintf("%s:%s", cfg.KafkaService.Host, cfg.KafkaService.Port)}
    pub := publisher.NewDefaultKafkaPublisher(brokers)
    sub := publisher.NewDefaultKafkaSubscriber(brokers)

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
	uc := usecase.NewDefaultOrderUsecase(
		orderRepo, 
		httpWalletHandler, 
		trafficUsecase, 
		bankDetailUsecase, 
		teamRelationsUsecase,
		pub,
		sub,
	)
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

	// Основной worker
	go uc.StartWorker(context.Background())

	// Retry worker
	go uc.StartRetryWorker(context.Background())
	
	// Мониторинг зависших ордеров
	go uc.StartStuckOrdersMonitor(context.Background())

    // Периодический запуск CancelExpiredOrders
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                if err := uc.CancelExpiredOrders(context.Background()); err != nil {
                    log.Printf("Auto-cancel error: %v", err)
                }
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

	log.Printf("gRPC server started on %s:%s\n", cfg.GRPCServer.Host, cfg.GRPCServer.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
