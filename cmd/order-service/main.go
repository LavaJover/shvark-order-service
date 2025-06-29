package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/client"
	"github.com/LavaJover/shvark-order-service/internal/config"
	"github.com/LavaJover/shvark-order-service/internal/delivery/grpcapi"
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/usdt"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/grpc"
	"github.com/joho/godotenv"
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
	orderKafkaPublisher := kafka.NewKafkaPublisher([]string{"localhost:9092"}, "order-events")
	disputeKafkaPublisher := kafka.NewKafkaPublisher([]string{"localhost:9092"}, "dispute-events")

	// Init order repo
	orderRepo := postgres.NewDefaultOrderRepository(db)
	// Init bank detail repo
	bankDetailRepo := postgres.NewDefaultBankDetailRepo(db)
	// Init traffic repo
	trafficRepo := postgres.NewDefaultTrafficRepository(db)

	// Init banking client
	bankingAddr := "localhost:50057"
	bankingClient, err := client.NewbankingClient(bankingAddr)
	if err != nil {
		log.Fatalf("failed to init banking client: %v\n", err)
	}

	// Init wallet handler
	httpWalletHandler, err := handlers.NewHTTPWalletHandler()
	if err != nil {
		log.Fatalf("failed to init wallet usecase")
	}

	// Init traffic usecase
	trafficUsecase := usecase.NewDefaultTrafficUsecase(trafficRepo)
	// Init order usecase
	uc := usecase.NewDefaultOrderUsecase(orderRepo, bankDetailRepo, bankingClient, httpWalletHandler, trafficUsecase, orderKafkaPublisher)

	// dispute
	disputeRepo := postgres.NewDefaultDisputeRepository(db)
	disputeUc := usecase.NewDefaultDisputeUsecase(
		disputeRepo,
		httpWalletHandler,
		orderRepo,
		trafficRepo,
		disputeKafkaPublisher,
	)

	// Creating gRPC server
	grpcServer := grpc.NewServer()
	orderHandler := grpcapi.NewOrderHandler(uc, disputeUc)
	trafficHandler := grpcapi.NewTrafficHandler(trafficUsecase)

	orderpb.RegisterOrderServiceServer(grpcServer, orderHandler)
	orderpb.RegisterTrafficServiceServer(grpcServer, trafficHandler)

	// Start
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil{
		log.Fatalf("failed to listen: %v", err)
	}

	// order auto-cancel
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
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
		ticker := time.NewTicker(5*time.Minute)
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

	fmt.Printf("gRPC server started on %s:%s\n", cfg.Host, cfg.Port)
	if err := grpcServer.Serve(lis); err != nil{
		log.Fatalf("failed to serve: %v\n", err)
	}
}