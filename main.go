package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Model for Espresso Ratio Record
type EspressoRecord struct {
	ID               uint    `gorm:"primaryKey"`
	Coffee           float64 `json:"coffee" binding:"required"`
	Water            float64 `json:"water" binding:"required"`
	Ratio            float64 `json:"ratio"`
	CoffeeBeansBrand string  `json:"coffee_beans_brand"`
	GrindSize        float64 `json:"grind_size"`
	TasteNodes       string  `json:"taste_nodes"`
	Picture          string  `json:"picture"` // Store the picture as a base64 string or URL
}

var DB *gorm.DB

func initDatabase() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	DB.AutoMigrate(&EspressoRecord{})
}

func main() {
	initDatabase()
	router := gin.Default()

	// Serve static files from the static directory
	router.Static("/static", "./static")

	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		var records []EspressoRecord
		DB.Order("id desc").Limit(3).Find(&records)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"records": records,
			"showAll": false,
		})
	})

	router.GET("/records", func(c *gin.Context) {
		var records []EspressoRecord
		DB.Order("id desc").Find(&records)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"records": records,
			"showAll": true,
		})
	})

	router.GET("/api/records", getRecords) // JSON API endpoint
	router.POST("/records", createRecord)

	router.Run(":8080")
}

// Keep the getRecords function for API calls
func getRecords(c *gin.Context) {
	var records []EspressoRecord
	DB.Find(&records)
	c.JSON(http.StatusOK, records)
}

func createRecord(c *gin.Context) {
	var record EspressoRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	record.Ratio = record.Coffee / record.Water
	DB.Create(&record)
	c.JSON(http.StatusOK, record)
}
