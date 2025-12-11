package database

import (
	"fmt"
	"log"

	"github.com/richard9219/3kstory/internal/config"
	"github.com/richard9219/3kstory/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	logLevel := logger.Info
	if cfg.Env == "production" {
		logLevel = logger.Error
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connected successfully")
	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Scene{},
		&models.AITask{},
	)
}
