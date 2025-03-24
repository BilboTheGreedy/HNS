package middleware

import (
	"net/http"

	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles authentication for web routes
type AuthMiddleware struct {
	userRepo repository.UserRepository
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo: userRepo,
	}
}

// RequireAuth middleware checks if user is authenticated
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		if !helpers.IsAuthenticated(c) {
			// Remember the requested URL for redirection after login
			c.Set("returnUrl", c.Request.URL.String())

			// Redirect to login page
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// If authenticated, load user data
		userID, exists := helpers.GetCurrentUserID(c)
		if !exists {
			// Invalid session, clear it
			helpers.ClearSession(c)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Get user from database to ensure it's still valid
		user, err := m.userRepo.GetByID(c.Request.Context(), userID)
		if err != nil || !user.IsActive {
			helpers.ClearSession(c)

			// Set alert message about account being inactive or not found
			if user != nil && !user.IsActive {
				helpers.SetAlert(c, "warning", "Your account has been deactivated. Please contact an administrator.")
			}

			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Store user object in context
		c.Set("user", user)

		c.Next()
	}
}

// RequireAdmin middleware checks if user is an admin
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if authenticated
		if !helpers.IsAuthenticated(c) {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Check if user is admin
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.HTML(http.StatusForbidden, "403.html", gin.H{
				"Title":       "Access Denied",
				"Message":     "You do not have permission to access this page",
				"CurrentYear": gin.H{},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoadUser middleware loads user data into context for templates
func (m *AuthMiddleware) LoadUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add user data from session to context
		helpers.AddUserDataToContext(c)
		c.Next()
	}
}
