package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"
	"github.com/himaatluri/brew-metrics/internal/config"
	"github.com/himaatluri/brew-metrics/internal/database"
)

// SessionMiddleware initializes the session store
func SessionMiddleware(cfg *config.Config) gin.HandlerFunc {
	db, err := database.GetDB().DB()
	if err != nil {
		panic("failed to get database connection: " + err.Error())
	}
	store, err := postgres.NewStore(
		db,
		[]byte(cfg.SessionConfig.SecretKey),
	)
	if err != nil {
		panic("failed to create session store: " + err.Error())
	}

	store.Options(sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   int(cfg.SessionConfig.SessionTimeout.Seconds()),
		Secure:   cfg.SessionConfig.Secure,
		HttpOnly: cfg.SessionConfig.HttpOnly,
		SameSite: http.SameSiteLaxMode,
	})

	return sessions.Sessions(cfg.SessionConfig.SessionName, store)
}

// AuthRequired middleware checks for valid session
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
