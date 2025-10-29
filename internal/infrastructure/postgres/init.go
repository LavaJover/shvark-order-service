package postgres

import (
	"log"

	"github.com/LavaJover/shvark-order-service/internal/config"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/engine"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/repository/antifraud/rules"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func MustInitDB(cfg *config.OrderConfig) *gorm.DB {
	dsn := cfg.OrderDB.Dsn
	
	// Создаем соединение с GORM
	db, err := gorm.Open(pg.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to init db: %v\n", err.Error())
	}

	// Автомиграция моделей
	db.AutoMigrate(
		&models.DeviceModel{}, 
		&models.TrafficModel{}, 
		&models.BankDetailModel{}, 
		&models.OrderModel{}, 
		&models.DisputeModel{}, 
		&models.TeamRelationshipModel{},
		&models.PaymentProcessingLog{},
		&rules.AntiFraudRule{},
		&engine.AntiFraudAuditLog{},
		&engine.UnlockAuditLog{},
	)

	// Применяем SQL миграции
	applySQLMigrations(db, dsn)

	return db
}

func applySQLMigrations(db *gorm.DB, dsn string) {
	// Получаем underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get sql.DB: %v", err)
		return
	}

	// Создаем мигратор
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Printf("Failed to create migration driver: %v", err)
		return
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/infrastructure/postgres/migrations",
		"postgres", 
		driver,
	)
	if err != nil {
		log.Printf("Failed to create migration instance: %v", err)
		return
	}

	// Применяем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("Failed to apply migrations: %v", err)
		return
	}

	log.Println("SQL migrations applied successfully")
}