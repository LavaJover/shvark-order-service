package main

import (
	"fmt"
	"log"
	"net"

	"github.com/LavaJover/shvark-order-service/internal/client"
	"github.com/LavaJover/shvark-order-service/internal/config"
	"github.com/LavaJover/shvark-order-service/internal/delivery/grpcapi"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
	"google.golang.org/grpc"
)

func main() {
	// Reading config
	cfg := config.MustLoad()
	// Init database
	db := postgres.MustInitDB(cfg)

	// Init order repo
	orderRepo := postgres.NewDefaultOrderRepository(db)
	// Init order usecase
	uc := usecase.NewDefaultOrderUsecase(orderRepo)

	// Init banking client
	bankingAddr := "localhost:50057"
	bankingClient, err := client.NewbankingClient(bankingAddr)
	if err != nil {
		log.Fatalf("failed to init banking client: %v\n", err)
	}

	// Creating gRPC server
	grpcServer := grpc.NewServer()
	orderHandler := grpcapi.NewOrderHandler(uc, bankingClient)

	orderpb.RegisterOrderServiceServer(grpcServer, orderHandler)

	// Start
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil{
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Printf("gRPC server started on %s:%s\n", cfg.Host, cfg.Port)
	if err := grpcServer.Serve(lis); err != nil{
		log.Fatalf("failed to serve: %v\n", err)
	}
}