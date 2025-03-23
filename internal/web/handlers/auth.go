package handlers

import (
	"net/http"
	"time"

	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	BaseHandler
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo repository.UserRepository, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		BaseHandler: BaseHandler{},
		userRepo:    userRepo,
		jwtManager:  jwtManager,
	}
}

// LoginPage renders the login page
func (h *AuthHandler) LoginPage(c *gin.Context) {
	// If already logged in, redirect to dashboard
	loggedIn := helpers.IsAuthenticated(c)
	if loggedIn {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	// Render login page directly
	c.HTML(http.StatusOK, "pages/login.html", gin.H{
		"Title":       "Login",
		"Alert":       helpers.GetAlert(c),
		"CurrentYear": time.Now().Year(),
	})
}

// Login handles the login form submission
func (h *AuthHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	log.Info().Str("username", username).Msg("Web UI Login attempt")

	// Get user by username
	user, err := h.userRepo.GetByUsername(c.Request.Context(), username)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("User not found during login")
		helpers.SetAlert(c, "danger", "Invalid username or password")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Check if user is active
	if !user.IsActive {
		helpers.SetAlert(c, "danger", "User account is inactive")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Invalid password")
		helpers.SetAlert(c, "danger", "Invalid username or password")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	log.Info().Str("username", username).Msg("Authentication successful, setting session")

	// Generate JWT token for API access
	var token string
	if h.jwtManager != nil {
		token, err = h.jwtManager.GenerateToken(user)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate JWT token")
			// Continue without token
		}
	}

	// Set user session using the helper function
	err = helpers.SetUserSession(c, user, token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save session")
		helpers.SetAlert(c, "danger", "Login failed due to session error")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Update last login time
	err = h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to update last login time")
		// Continue regardless
	}

	log.Info().Str("username", username).Msg("Session saved, redirecting to dashboard")
	c.Redirect(http.StatusFound, "/dashboard")
}

// RegisterPage renders the registration page
func (h *AuthHandler) RegisterPage(c *gin.Context) {
	// If already logged in, redirect to home
	loggedIn := helpers.IsAuthenticated(c)
	if loggedIn {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	// Render register page
	c.HTML(http.StatusOK, "pages/register.html", gin.H{
		"Title":       "Register",
		"Alert":       helpers.GetAlert(c),
		"CurrentYear": time.Now().Year(),
	})
}

// Register handles the registration form submission
func (h *AuthHandler) Register(c *gin.Context) {
	// Get form data
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")

	// Validate passwords match
	if password != confirmPassword {
		helpers.SetAlert(c, "danger", "Passwords do not match")
		c.Redirect(http.StatusFound, "/register")
		return
	}

	// Check if username already exists
	existingUser, err := h.userRepo.GetByUsername(c.Request.Context(), username)
	if err == nil && existingUser != nil {
		helpers.SetAlert(c, "danger", "Username already exists")
		c.Redirect(http.StatusFound, "/register")
		return
	}

	// Check if email already exists
	existingUser, err = h.userRepo.GetByEmail(c.Request.Context(), email)
	if err == nil && existingUser != nil {
		helpers.SetAlert(c, "danger", "Email already exists")
		c.Redirect(http.StatusFound, "/register")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		helpers.SetAlert(c, "danger", "Registration failed: internal error")
		c.Redirect(http.StatusFound, "/register")
		return
	}

	// Create user
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Role:         models.RoleUser,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save user
	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		helpers.SetAlert(c, "danger", "Registration failed: "+err.Error())
		c.Redirect(http.StatusFound, "/register")
		return
	}

	// Success message and redirect to login
	helpers.SetAlert(c, "success", "Registration successful! You can now log in.")
	c.Redirect(http.StatusFound, "/login")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear session
	helpers.ClearSession(c)

	// Redirect to login
	c.Redirect(http.StatusFound, "/login")
}
