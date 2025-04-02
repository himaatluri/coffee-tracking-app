package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/himaatluri/brew-metrics/internal/config"
	"github.com/himaatluri/brew-metrics/internal/database"
	"github.com/himaatluri/brew-metrics/internal/handlers"
)

func main() {
	// Initialize configuration
	cfg := config.DefaultConfig()

	// Initialize database
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize router
	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob("templates/*")

	// Serve static files
	r.Static("/static", "./static")

	// Routes
	r.GET("/", handlers.GetHomePage)
	r.GET("/records", handlers.GetAllRecords)
	r.GET("/api/records", handlers.GetRecords)
	r.POST("/records", handlers.CreateRecord)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
