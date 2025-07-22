package config

import (
	"log"
	"os"
	"github.com/ilyakaznacheev/cleanenv"
)

type OrderConfig struct {
	Env string 	   `yaml:"env"`
	GRPCServer 	   `yaml:"grpc_server"`
	OrderDB 	   `yaml:"order_db"`
	LogConfig 	   `yaml:"log_config"`
	BankingService `yaml:"banking-service"`
	WalletService  `yaml:"wallet-service"`
	KafkaService   `yaml:"kafka-service"`
}

type KafkaService struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type GRPCServer struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type OrderDB struct {
	Dsn string `yaml:"dsn"`
}

type LogConfig struct {
	LogLevel 	string 	`yaml:"log_level"`
	LogFormat 	string 	`yaml:"log_format"`
	LogOutput 	string 	`yaml:"log_output"`
}

type BankingService struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type WalletService struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func MustLoad() *OrderConfig {

	// Processing env config variable and file
	configPath := os.Getenv("ORDER_CONFIG_PATH")

	if configPath == ""{
		log.Fatalf("ORDER_CONFIG_PATH was not found\n")
	}

	if _, err := os.Stat(configPath); err != nil{
		log.Fatalf("failed to find config file: %v\n", err)
	}

	// YAML to struct object
	var cfg OrderConfig
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil{
		log.Fatalf("failed to read config file: %v", err)
	}

	return &cfg
}