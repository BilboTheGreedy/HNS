package middleware

import (
	"net/http"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware contains auth middleware functions
type AuthMiddleware struct {
	userRepo repository.UserRepository
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo: userRepo,
	}
}

// AuthRequired middleware checks if user is authenticated
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(helpers.SessionKeyUserID)

		// Log session data for debugging
		log.Debug().
			Interface("userID", userID).
			Interface("username", session.Get(helpers.SessionKeyUsername)).
			Interface("loggedIn", session.Get(helpers.SessionKeyLoggedIn)).
			Str("path", c.Request.URL.Path).
			Msg("Auth check - session data")

		if userID == nil {
			log.Warn().Str("path", c.Request.URL.Path).Msg("No userID in session, redirecting to login")
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Get user data for templates
		user, err := m.userRepo.GetByID(c.Request.Context(), userID.(int64))
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user by ID")
			session.Clear()
			session.Save()
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Make user data available in templates
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("username", user.Username)
		c.Set("isAdmin", user.Role == models.RoleAdmin)
		c.Set("loggedIn", true)

		c.Next()
	}
}

// AdminRequired middleware checks if the user is an admin
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if user.(*models.User).Role != models.RoleAdmin {
			c.HTML(http.StatusForbidden, "pages/403.html", gin.H{
				"Title": "Access Denied",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoggerMiddleware logs incoming requests and their responses
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Log minimal request info
		if c.Writer.Status() >= 400 {
			// Log errors with more detail
			log.Warn().
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Int("status", c.Writer.Status()).
				Int("size", c.Writer.Size()).
				Int("errors", len(c.Errors)).
				Msg("Request error")
		}
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RecoveryMiddleware recovers from panics and logs the error
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Interface("error", err).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Msg("Panic recovered")

				c.HTML(http.StatusInternalServerError, "pages/500.html", gin.H{
					"Title": "Internal Server Error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
