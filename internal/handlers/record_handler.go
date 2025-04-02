package handlers

import (
	"net/http"

	"github.com/himaatluri/brew-metrics/internal/database"
	"github.com/himaatluri/brew-metrics/internal/models"

	"github.com/gin-gonic/gin"
)

// GetHomePage renders the home page with recent records
func GetHomePage(c *gin.Context) {
	var records []models.EspressoRecord
	result := database.DB.Order("created_at desc").Limit(5).Find(&records)
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "index.html", gin.H{
			"error":   "Failed to load records",
			"showAll": false,
		})
		return
	}

	// Get unique coffee brands
	var brands []string
	database.DB.Model(&models.EspressoRecord{}).Distinct().Pluck("beans_brand", &brands)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"records": records,
		"brands":  brands,
		"showAll": false,
	})
}

// GetAllRecords renders the records page with all records
func GetAllRecords(c *gin.Context) {
	var records []models.EspressoRecord
	result := database.DB.Order("created_at desc").Find(&records)
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "index.html", gin.H{
			"error":   "Failed to load records",
			"showAll": true,
		})
		return
	}

	// Get unique coffee brands
	var brands []string
	database.DB.Model(&models.EspressoRecord{}).Distinct().Pluck("beans_brand", &brands)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"records": records,
		"brands":  brands,
		"showAll": true,
	})
}

// GetRecords returns all records as JSON
func GetRecords(c *gin.Context) {
	var records []models.EspressoRecord
	result := database.DB.Order("created_at desc").Find(&records)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load records"})
		return
	}

	c.JSON(http.StatusOK, records)
}

// CreateRecord handles the creation of a new coffee record
func CreateRecord(c *gin.Context) {
	var record models.EspressoRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate ratio before saving
	if record.Coffee > 0 {
		record.Ratio = record.Water / record.Coffee
	}

	result := database.DB.Create(&record)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create record"})
		return
	}

	c.JSON(http.StatusCreated, record)
}
