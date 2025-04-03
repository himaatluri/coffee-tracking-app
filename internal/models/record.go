package models

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// EspressoRecord represents a coffee brewing record with measurements and metadata
type EspressoRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Coffee     float64   `json:"coffee" binding:"required"`
	Water      float64   `json:"water" gorm:"default:36"` // Default to 36g water (1:2 ratio with 18g coffee)
	Ratio      float64   `json:"ratio"`
	BeansBrand string    `json:"beans_brand"`
	GrindSize  float64   `json:"grind_size"`
	TasteNodes string    `json:"taste_nodes"`
	Picture    string    `json:"picture" gorm:"type:text"` // Base64 encoded image
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ValidatePicture checks if the picture is a valid base64 encoded image
func (r *EspressoRecord) ValidatePicture() bool {
	if r.Picture == "" {
		return true
	}

	// Check if it's a valid base64 string
	parts := strings.Split(r.Picture, ",")
	if len(parts) != 2 {
		return false
	}

	// Try to decode the base64 string
	_, err := base64.StdEncoding.DecodeString(parts[1])
	return err == nil
}

// CalculateRatio computes the water-to-coffee ratio
func (r *EspressoRecord) CalculateRatio() {
	if r.Coffee > 0 && r.Water > 0 {
		r.Ratio = r.Water / r.Coffee
	} else if r.Coffee > 0 {
		// If water is not specified, use default 1:2 ratio
		r.Water = r.Coffee * 2
		r.Ratio = 2
	}
}

// BeforeCreate is a GORM hook that runs before creating a record
func (r *EspressoRecord) BeforeCreate() error {
	r.CalculateRatio()

	// Validate picture if present
	if r.Picture != "" && !r.ValidatePicture() {
		return fmt.Errorf("invalid picture format")
	}

	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a record
func (r *EspressoRecord) BeforeUpdate() error {
	r.CalculateRatio()

	// Validate picture if present
	if r.Picture != "" && !r.ValidatePicture() {
		return fmt.Errorf("invalid picture format")
	}

	r.UpdatedAt = time.Now()
	return nil
}
