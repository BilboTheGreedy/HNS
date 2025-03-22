package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// SessionKeys holds the key names used in the session
const (
	SessionKeyUserID    = "userID"
	SessionKeyUsername  = "username"
	SessionKeyIsAdmin   = "isAdmin"
	SessionKeyLoggedIn  = "loggedIn"
	SessionKeyToken     = "token"
	SessionKeyAlertType = "alertType"
	SessionKeyAlertMsg  = "alertMessage"
)

// SetupSessionStore configures session management for the application
func SetupSessionStore(router *gin.Engine) {
	// Generate secure keys for cookies
	authKey := generateKey(32) // 32 bytes for AES-256
	encKey := generateKey(32)  // 32 bytes for encryption

	// Log keys in development (remove in production)
	log.Debug().Str("auth_key", base64.StdEncoding.EncodeToString(authKey)).Msg("Auth key generated")

	// Create cookie store with secure keys
	store := cookie.NewStore(authKey, encKey)

	// Configure session options
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1 day
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Register the session middleware
	router.Use(sessions.Sessions("hns_session", store))
}

// generateKey creates a random key of the specified length
func generateKey(length int) []byte {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		// If random generation fails, use a fallback key (in dev only)
		log.Error().Err(err).Msg("Failed to generate random key, using fallback")
		// Warning: Use environment variables in production
		if length == 32 {
			return []byte("01234567890123456789012345678901")
		}
		return []byte("0123456789012345")
	}
	return key
}

// SetUserSession stores user information in the session
func SetUserSession(c *gin.Context, user *models.User, token string) error {
	session := sessions.Default(c)

	// Clear any existing session data
	session.Clear()

	// Set session data
	session.Set(SessionKeyUserID, user.ID)
	session.Set(SessionKeyUsername, user.Username)
	session.Set(SessionKeyIsAdmin, user.Role == models.RoleAdmin)
	session.Set(SessionKeyLoggedIn, true)

	// Store JWT token if provided
	if token != "" {
		session.Set(SessionKeyToken, token)
	}

	// Save session immediately
	err := session.Save()
	if err != nil {
		log.Error().Err(err).Msg("Failed to save user session")
		return err
	}

	return nil
}

// GetUserFromSession retrieves user information from the session
func GetUserFromSession(c *gin.Context) *models.User {
	session := sessions.Default(c)
	userID := session.Get(SessionKeyUserID)

	if userID == nil {
		return nil
	}

	// Create a user object with session data
	user := &models.User{
		ID:       userID.(int64),
		Username: session.Get(SessionKeyUsername).(string),
	}

	// Set role if available
	isAdmin := session.Get(SessionKeyIsAdmin)
	if isAdmin != nil && isAdmin.(bool) {
		user.Role = models.RoleAdmin
	} else {
		user.Role = models.RoleUser
	}

	return user
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	session := sessions.Default(c)
	userID := session.Get(SessionKeyUserID)
	return userID != nil
}

// IsAdmin checks if the user is an admin
func IsAdmin(c *gin.Context) bool {
	session := sessions.Default(c)
	isAdmin := session.Get(SessionKeyIsAdmin)
	return isAdmin != nil && isAdmin.(bool)
}

// ClearSession removes all session data
func ClearSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
}

// AddContextUserData adds user information to the Gin context
func AddContextUserData(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get(SessionKeyUserID)
	username := session.Get(SessionKeyUsername)
	isAdmin := session.Get(SessionKeyIsAdmin)
	loggedIn := session.Get(SessionKeyLoggedIn)

	if userID != nil {
		c.Set(SessionKeyUserID, userID)
	}

	if username != nil {
		c.Set(SessionKeyUsername, username)
	}

	if isAdmin != nil {
		c.Set(SessionKeyIsAdmin, isAdmin)
	}

	if loggedIn != nil {
		c.Set(SessionKeyLoggedIn, loggedIn)
	}
}
