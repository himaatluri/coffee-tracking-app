package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	secretLength = 32
	secretFile   = ".jwt_secret"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"unique" json:"email"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Auth struct {
	db          *gorm.DB
	jwtSecret   []byte
	tokenExpiry time.Duration
}

func NewAuth(db *gorm.DB) *Auth {
	auth := &Auth{
		db:          db,
		tokenExpiry: 24 * time.Hour,
	}
	return auth
}

func (a *Auth) Initialize() error {
	if err := a.loadOrGenerateSecret(); err != nil {
		return fmt.Errorf("failed to initialize JWT secret: %v", err)
	}
	return a.db.AutoMigrate(&User{})
}

func (a *Auth) RegisterUser(email, password string) error {
	// Validate input
	if err := validateEmail(email); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}

	// Check if user already exists
	var existingUser User
	if result := a.db.Where("email = ?", email).First(&existingUser); result.Error == nil {
		return fmt.Errorf("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := User{
		Email:    email,
		Password: string(hashedPassword),
	}

	result := a.db.Create(&user)
	return result.Error
}

func (a *Auth) LoginUser(email, password string) (string, error) {
	var user User
	if err := a.db.Where("email = ?", email).First(&user).Error; err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(a.tokenExpiry).Unix(),
	})

	return token.SignedString(a.jwtSecret)
}

func (a *Auth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for login, signup, and static files
		if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/signup" ||
			strings.HasPrefix(c.Request.URL.Path, "/static") || c.Request.URL.Path == "/logout" {
			c.Next()
			return
		}

		// Get token from Authorization header
		token := c.GetHeader("Authorization")
		if token == "" {
			// If no token in header, check query string
			token = c.Query("token")
		}

		if token == "" {
			// For HTML requests, redirect to login with return URL
			if strings.HasPrefix(c.GetHeader("Accept"), "text/html") {
				returnURL := c.Request.URL.Path
				if c.Request.URL.RawQuery != "" {
					returnURL += "?" + c.Request.URL.RawQuery
				}
				c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/login?return=%s", url.QueryEscape(returnURL)))
				return
			}
			// For API requests, return 401
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			return
		}

		// Remove "Bearer " prefix if present
		token = strings.TrimPrefix(token, "Bearer ")

		// Validate token
		claims, err := a.validateToken(token)
		if err != nil {
			// For HTML requests, redirect to login with return URL
			if strings.HasPrefix(c.GetHeader("Accept"), "text/html") {
				returnURL := c.Request.URL.Path
				if c.Request.URL.RawQuery != "" {
					returnURL += "?" + c.Request.URL.RawQuery
				}
				c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/login?return=%s", url.QueryEscape(returnURL)))
				return
			}
			// For API requests, return 401
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Set user ID in context
		userID := uint((*claims)["user_id"].(float64))
		c.Set("user_id", userID)

		// For HTML requests, ensure token is in context for templates
		if strings.HasPrefix(c.GetHeader("Accept"), "text/html") {
			c.Set("auth_token", token)
		}

		c.Next()
	}
}

func generateSecret() ([]byte, error) {
	secret := make([]byte, secretLength)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	return secret, nil
}

func (a *Auth) loadOrGenerateSecret() error {
	// Check if secret file exists
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	secretPath := filepath.Join(home, secretFile)
	secret, err := os.ReadFile(secretPath)

	if os.IsNotExist(err) {
		// Generate new secret
		secret, err = generateSecret()
		if err != nil {
			return err
		}

		// Save secret with restricted permissions
		err = os.WriteFile(secretPath, secret, 0600)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	a.jwtSecret = secret
	return nil
}

func (a *Auth) validateToken(token string) (*jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token is expired
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return nil, errors.New("token expired")
		}
	} else {
		return nil, errors.New("invalid token expiration")
	}

	return &claims, nil
}
