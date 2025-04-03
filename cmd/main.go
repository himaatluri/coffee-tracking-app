package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/himaatluri/brew-metrics/internal/config"
	"github.com/himaatluri/brew-metrics/internal/database"
	"github.com/himaatluri/brew-metrics/internal/handlers"
)

func main() {
	// Load configuration
	cfg := config.DefaultConfig()

	// Initialize database
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Routes
	r.GET("/", handlers.GetHomePage)
	r.GET("/records", handlers.GetAllRecords)
	r.POST("/records", handlers.CreateRecord)
	r.PUT("/records/:id", handlers.UpdateRecord)
	r.DELETE("/records/:id", handlers.DeleteRecord)

	r.Run(cfg.Port)
}
