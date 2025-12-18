package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/LavaJover/shvark-order-service/internal/app/background"
	"github.com/LavaJover/shvark-order-service/internal/app/setup"
	"github.com/LavaJover/shvark-order-service/internal/delivery/grpcapi"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen/order"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("Note: .env file not found, using environment variables")
    }

    deps, err := setup.InitializeDependencies()
    if err != nil {
        log.Fatalf("Failed to initialize dependencies: %v", err)
    }

    useCases, err := setup.InitializeUseCases(deps)
    if err != nil {
        log.Fatalf("Failed to initialize use cases: %v", err)
    }

    // Инициализация антифрода (теперь отдельно)
    antiFraudSystem, err := setup.InitializeAntiFraud(deps)
    if err != nil {
        log.Fatalf("Failed to initialize anti-fraud system: %v", err)
    }

    // Создание и запуск gRPC сервера
    grpcServer := setupGRPCServer(useCases, antiFraudSystem)
    
    // Запуск фоновых задач
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    bgTasks := background.NewBackgroundTasks(
        useCases.OrderUsecase,
        useCases.DisputeUsecase, 
        useCases.DeviceUsecase,
    )
    bgTasks.StartAll(ctx)
    
    // Запуск планировщика антифрода
    go antiFraudSystem.Scheduler.Start(ctx)

    // Запуск gRPC сервера
    lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", deps.Config.GRPCServer.Host, deps.Config.GRPCServer.Port))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    log.Printf("gRPC server starting on %s:%s", deps.Config.GRPCServer.Host, deps.Config.GRPCServer.Port)
    
    // Обработка graceful shutdown
    go gracefulShutdown(grpcServer, cancel)

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }

    go func() {
        http.Handle("/metrics", promhttp.Handler())
        http.ListenAndServe(":8081", nil)
    }()
}

func setupGRPCServer(useCases *setup.UseCases, antiFraudSystem *setup.AntiFraudSystem) *grpc.Server {
    server := grpc.NewServer()
    
    // Регистрация всех обработчиков
    orderpb.RegisterOrderServiceServer(server, 
        grpcapi.NewOrderHandler(
            useCases.OrderUsecase, 
            useCases.DisputeUsecase, 
            useCases.BankDetailUsecase, 
            useCases.AutomaticUsecase,
        ))
    
    orderpb.RegisterTrafficServiceServer(server, 
        grpcapi.NewTrafficHandler(useCases.TrafficUsecase))
    
    orderpb.RegisterBankDetailServiceServer(server, 
        grpcapi.NewBankDetailHandler(useCases.BankDetailUsecase))
    
    orderpb.RegisterTeamRelationsServiceServer(server, 
        grpcapi.NewTeamRelationsHandler(useCases.TeamRelationsUsecase))
    
    orderpb.RegisterDeviceServiceServer(server, 
        grpcapi.NewDeviceHandler(useCases.DeviceUsecase))
    
    // Используем antiFraudSystem.UseCase вместо useCases.AntiFraudUseCase
    orderpb.RegisterAntiFraudServiceServer(server, 
        grpcapi.NewAntiFraudHandler(antiFraudSystem.UseCase))
    
    return server
}

func gracefulShutdown(server *grpc.Server, cancel context.CancelFunc) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    <-sigChan
    log.Println("Shutdown signal received, stopping server...")
    
    cancel()
    server.GracefulStop()
    log.Println("Server stopped gracefully")
}