package handlers

import (
	"net/http"

	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	BaseHandler
	userRepo repository.UserRepository
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		BaseHandler: *NewBaseHandler(),
		userRepo:    userRepo,
	}
}

// ShowLogin shows the login page
func (h *AuthHandler) ShowLogin(c *gin.Context) {
	// If user is already logged in, redirect to dashboard
	if helpers.IsAuthenticated(c) {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	// Check if there's a return URL in the query parameters
	returnURL := c.Query("returnUrl")
	if returnURL != "" {
		c.Set("returnUrl", returnURL)
	}

	h.RenderTemplate(c, "login", gin.H{
		"Title": "Login",
	})
}

// Login handles the login form submission
func (h *AuthHandler) Login(c *gin.Context) {
	// Get form data
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Validate input
	if username == "" || password == "" {
		h.RedirectWithAlert(c, "/login", "danger", "Username and password are required")
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByUsername(c.Request.Context(), username)
	if err != nil {
		h.RedirectWithAlert(c, "/login", "danger", "Invalid username or password")
		return
	}

	// Check if user is active
	if !user.IsActive {
		h.RedirectWithAlert(c, "/login", "warning", "Your account is inactive. Please contact an administrator.")
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		h.RedirectWithAlert(c, "/login", "danger", "Invalid username or password")
		return
	}

	// Create session
	helpers.SetUserSession(c, user)

	// Update last login time
	if err := h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID); err != nil {
		// Log error but don't stop the login process
	}

	// Get return URL if any
	returnURL, exists := c.Get("returnUrl")
	if exists && returnURL.(string) != "" {
		c.Redirect(http.StatusFound, returnURL.(string))
		return
	}

	// Redirect to dashboard
	c.Redirect(http.StatusFound, "/dashboard")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear session
	helpers.ClearSession(c)

	// Redirect to login page
	h.RedirectWithAlert(c, "/login", "success", "You have been logged out successfully")
}
