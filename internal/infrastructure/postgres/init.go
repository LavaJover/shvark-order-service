package postgres

import (
	"log"

	"github.com/LavaJover/shvark-order-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func MustInitDB(cfg *config.OrderConfig) *gorm.DB {
	dsn := cfg.OrderDB.Dsn
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to init db: %v\n", err.Error())
	}

	db.AutoMigrate(&TrafficModel{}, &BankDetailModel{}, &OrderModel{})

	return db
}