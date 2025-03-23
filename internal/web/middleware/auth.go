package middleware

import (
	"net/http"

	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
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
		// Check if user is authenticated
		loggedIn, exists := c.Get("loggedIn")
		userID, userExists := c.Get("userID")

		log.Debug().
			Interface("loggedIn", loggedIn).
			Interface("exists", exists).
			Interface("userID", userID).
			Interface("userExists", userExists).
			Str("path", c.Request.URL.Path).
			Msg("Auth check - session data")

		if !exists || !loggedIn.(bool) || !userExists {
			log.Info().Str("path", c.Request.URL.Path).Msg("No authenticated user, redirecting to login")
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Get user data for templates
		user, err := m.userRepo.GetByID(c.Request.Context(), userID.(int64))
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user by ID")
			helpers.ClearSession(c)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Make user data available in templates
		c.Set("user", user)

		c.Next()
	}
}

// AdminRequired middleware checks if the user is an admin
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.HTML(http.StatusForbidden, "pages/403.html", gin.H{
				"Title": "Access Denied",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
