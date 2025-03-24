package helpers

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// Session key constants
const (
	UserIDKey    = "userID"
	UsernameKey  = "username"
	IsAdminKey   = "isAdmin"
	LoggedInKey  = "loggedIn"
	AlertTypeKey = "alertType"
	AlertMsgKey  = "alertMessage"
)

// Alert represents a flash message to show to the user
type Alert struct {
	Type    string // success, danger, warning, info
	Message string
}

// SetupSessionStore initializes the session store for the application
func SetupSessionStore(router *gin.Engine) {
	// Generate random key for session encryption
	key := generateRandomKey(32) // 32 bytes for AES-256

	// Create cookie store
	store := cookie.NewStore(key)

	// Configure session
	store.Options(sessions.Options{
		Path:     "/",   // Available to all paths
		MaxAge:   86400, // 1 day
		HttpOnly: true,  // Not accessible via JavaScript
		Secure:   false, // Set to true in production if using HTTPS
	})

	// Register the session middleware
	router.Use(sessions.Sessions("hns_session", store))
}

// generateRandomKey generates a random key for session encryption
func generateRandomKey(length int) []byte {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		// If random generation fails, use a default key
		// This is not secure for production; consider using env variables
		return []byte("default-key-please-change-in-production")
	}
	return key
}

// SetUserSession stores user information in the session
func SetUserSession(c *gin.Context, user *models.User) {
	session := sessions.Default(c)

	// Clear any existing session data
	session.Clear()

	// Store user data in session
	session.Set(UserIDKey, user.ID)
	session.Set(UsernameKey, user.Username)
	session.Set(IsAdminKey, user.Role == models.RoleAdmin)
	session.Set(LoggedInKey, true)

	// Save the session
	session.Save()
}

// ClearSession removes all session data (logout)
func ClearSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	session := sessions.Default(c)
	return session.Get(LoggedInKey) == true
}

// GetCurrentUserID gets the current user's ID from the session
func GetCurrentUserID(c *gin.Context) (int64, bool) {
	session := sessions.Default(c)
	userID := session.Get(UserIDKey)
	if userID == nil {
		return 0, false
	}
	return userID.(int64), true
}

// SetAlert sets a flash message in the session
func SetAlert(c *gin.Context, alertType, message string) {
	session := sessions.Default(c)
	session.Set(AlertTypeKey, alertType)
	session.Set(AlertMsgKey, message)
	session.Save()
}

// GetAlert gets and clears the flash message from the session
func GetAlert(c *gin.Context) *Alert {
	session := sessions.Default(c)

	alertType := session.Get(AlertTypeKey)
	alertMsg := session.Get(AlertMsgKey)

	if alertType != nil && alertMsg != nil {
		// Clear the alert after retrieving it
		session.Delete(AlertTypeKey)
		session.Delete(AlertMsgKey)
		session.Save()

		return &Alert{
			Type:    alertType.(string),
			Message: alertMsg.(string),
		}
	}

	return nil
}

// AddUserDataToContext adds user data from session to the Gin context
func AddUserDataToContext(c *gin.Context) {
	session := sessions.Default(c)

	// Get user data from session
	userID := session.Get(UserIDKey)
	username := session.Get(UsernameKey)
	isAdmin := session.Get(IsAdminKey)
	loggedIn := session.Get(LoggedInKey)

	// Add to context if available
	if userID != nil {
		c.Set(UserIDKey, userID)
	}

	if username != nil {
		c.Set(UsernameKey, username)
	}

	if isAdmin != nil {
		c.Set(IsAdminKey, isAdmin)
	}

	if loggedIn != nil {
		c.Set(LoggedInKey, loggedIn)
	}
}

// GenerateRandomToken generates a random token for CSRF protection
func GenerateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
