package postgres

import (
	"log"

	"github.com/LavaJover/shvark-order-service/internal/config"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func MustInitDB(cfg *config.OrderConfig) *gorm.DB {
	dsn := cfg.OrderDB.Dsn
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to init db: %v\n", err.Error())
	}

	db.AutoMigrate(&models.TrafficModel{}, &models.BankDetailModel{}, &models.OrderModel{}, &models.DisputeModel{}, &models.TeamRelationshipModel{})

	return db
}