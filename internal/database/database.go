package database

import (
	"fmt"
	"log"

	"github.com/himaatluri/brew-metrics/internal/config"
	"github.com/himaatluri/brew-metrics/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the database connection and runs migrations
func InitDB(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
		cfg.DBConfig.User,
		cfg.DBConfig.Password,
		cfg.DBConfig.DBName,
		cfg.DBConfig.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Run migrations
	err = DB.AutoMigrate(&models.EspressoRecord{})
	if err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// GetDB returns the initialized database connection
func GetDB() *gorm.DB {
	if DB == nil {
		panic("database not initialized - call InitDB first")
	}
	return DB
}
