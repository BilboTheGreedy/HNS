package handlers

import (
	"net/http"

	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo repository.UserRepository, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// LoginPage renders the login page
func (h *AuthHandler) LoginPage(c *gin.Context) {
	// If already logged in, redirect to home
	session := sessions.Default(c)
	if session.Get(helpers.SessionKeyUserID) != nil {
		c.Redirect(http.StatusFound, "/hostnames")
		return
	}

	// Render login page
	c.HTML(http.StatusOK, "pages/login.html", gin.H{
		"Title":       "Login",
		"Alert":       helpers.GetAlert(c),
		"CurrentYear": helpers.GetCurrentYear(),
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

	// Set user session
	session := sessions.Default(c)
	session.Set(helpers.SessionKeyUserID, user.ID)
	session.Set(helpers.SessionKeyUsername, user.Username)
	session.Set(helpers.SessionKeyIsAdmin, user.Role == models.RoleAdmin)
	session.Set(helpers.SessionKeyLoggedIn, true)

	// Set token if generated
	if token != "" {
		session.Set(helpers.SessionKeyToken, token)
	}

	err = session.Save()
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

	log.Info().Str("username", username).Msg("Session saved, redirecting to hostnames page")
	c.Redirect(http.StatusFound, "/hostnames")
}

// RegisterPage renders the registration page
func (h *AuthHandler) RegisterPage(c *gin.Context) {
	// If already logged in, redirect to home
	session := sessions.Default(c)
	if session.Get(helpers.SessionKeyUserID) != nil {
		c.Redirect(http.StatusFound, "/hostnames")
		return
	}

	// Render register page
	c.HTML(http.StatusOK, "pages/register.html", gin.H{
		"Title":       "Register",
		"Alert":       helpers.GetAlert(c),
		"CurrentYear": helpers.GetCurrentYear(),
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
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	// Redirect to login
	c.Redirect(http.StatusFound, "/login")
}

// ChangePassword handles password change requests
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SetAlert(c, "danger", "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get current user
	user, err := h.userRepo.GetByID(c.Request.Context(), userID.(int64))
	if err != nil {
		helpers.SetAlert(c, "danger", "User not found")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Get form data
	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		helpers.SetAlert(c, "danger", "Current password is incorrect")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Validate passwords match
	if newPassword != confirmPassword {
		helpers.SetAlert(c, "danger", "New passwords do not match")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		helpers.SetAlert(c, "danger", "Failed to update password")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to update password")
		helpers.SetAlert(c, "danger", "Failed to update password: "+err.Error())
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Success
	helpers.SetAlert(c, "success", "Password updated successfully")
	c.Redirect(http.StatusFound, "/profile")
}

// ForgotPassword handles password reset requests
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	// This is a placeholder for a real implementation
	// In a real application, this would send a reset email

	email := c.PostForm("email")
	if email == "" {
		helpers.SetAlert(c, "danger", "Email is required")
		c.Redirect(http.StatusFound, "/forgot-password")
		return
	}

	// Check if user exists
	user, err := h.userRepo.GetByEmail(c.Request.Context(), email)
	if err != nil || user == nil {
		// Don't reveal if email exists for security
		helpers.SetAlert(c, "success", "If your email is registered, you will receive a password reset link")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// In a real implementation, generate a token and send email
	// For now, just show a success message
	helpers.SetAlert(c, "success", "If your email is registered, you will receive a password reset link")
	c.Redirect(http.StatusFound, "/login")
}
