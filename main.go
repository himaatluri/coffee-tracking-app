package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/himaatluri/coffee-tracking-app/pkg/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Model for Espresso Ratio Record
type EspressoRecord struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	User       auth.User `gorm:"foreignKey:UserID" json:"-"`
	Coffee     float64   `json:"coffee" binding:"required"`
	Water      float64   `json:"water" binding:"required"`
	Ratio      float64   `json:"ratio"`
	BeansBrand string    `json:"beans_brand"`
	GrindSize  float64   `json:"grind_size"`
	TasteNodes string    `json:"taste_nodes"`
	Picture    string    `json:"picture"` // Store the picture as a base64 string or URL
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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

	// Drop the table if it exists to ensure clean migration
	DB.Migrator().DropTable(&EspressoRecord{})

	// Run the migration
	if err := DB.AutoMigrate(&EspressoRecord{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed successfully")
}

func main() {
	initDatabase()

	// Initialize auth without JWT secret parameter
	authHandler := auth.NewAuth(DB)
	if err := authHandler.Initialize(); err != nil {
		log.Fatalf("Failed to initialize auth: %v", err)
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	// Auth routes
	router.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", nil)
	})

	router.GET("/login", func(c *gin.Context) {
		registered := c.Query("registered")
		c.HTML(http.StatusOK, "login.html", gin.H{
			"registered": registered == "true",
		})
	})

	router.GET("/logout", func(c *gin.Context) {
		// Clear any server-side session data if needed
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	})

	router.POST("/login", func(c *gin.Context) {
		var login struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&login); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		token, err := authHandler.LoginUser(login.Email, login.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Set success response with token
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"token":   token,
			"message": "Login successful",
		})
	})

	router.POST("/register", func(c *gin.Context) {
		var register struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&register); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := authHandler.RegisterUser(register.Email, register.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	})

	// Protected routes
	protected := router.Group("/")
	protected.Use(authHandler.AuthMiddleware())
	{
		protected.GET("/", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			var records []EspressoRecord
			DB.Where("user_id = ?", userID).Order("id desc").Limit(3).Find(&records)
			c.HTML(http.StatusOK, "index.html", gin.H{
				"records": records,
				"showAll": false,
			})
		})

		protected.GET("/records", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			var records []EspressoRecord
			DB.Where("user_id = ?", userID).Order("id desc").Find(&records)
			c.HTML(http.StatusOK, "index.html", gin.H{
				"records": records,
				"showAll": true,
			})
		})

		protected.GET("/api/records", getRecords)
		protected.POST("/records", createRecord)
	}

	router.Run(":8080")
}

// Keep the getRecords function for API calls
func getRecords(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var records []EspressoRecord
	if err := DB.Where("user_id = ?", userID).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records"})
		return
	}
	c.JSON(http.StatusOK, records)
}

func createRecord(c *gin.Context) {
	var record EspressoRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	record.UserID = userID.(uint)
	record.Ratio = record.Coffee / record.Water

	if err := DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create record"})
		return
	}

	c.JSON(http.StatusOK, record)
}
