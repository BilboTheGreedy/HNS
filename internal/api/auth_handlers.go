package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo      repository.UserRepository
	jwtManager    *auth.JWTManager
	apiKeyManager *auth.APIKeyManager
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo repository.UserRepository, jwtManager *auth.JWTManager, apiKeyManager *auth.APIKeyManager) *AuthHandler {
	return &AuthHandler{
		userRepo:      userRepo,
		jwtManager:    jwtManager,
		apiKeyManager: apiKeyManager,
	}
}

// RegisterUser handles user registration requests
func (h *AuthHandler) RegisterUser(c *gin.Context) {
	// Parse request
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username already exists
	existingUser, err := h.userRepo.GetByUsername(c.Request.Context(), req.Username)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	// Check if email already exists
	existingUser, err = h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		log.Error().Err(err).Msg("Failed to hash password")
		return
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         models.Role(req.Role),
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save user
	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		log.Error().Err(err).Msg("Failed to create user")
		return
	}

	// Remove password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, user)
}

// Login handles user login requests
func (h *AuthHandler) Login(c *gin.Context) {
	// Parse request
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by username
	user, err := h.userRepo.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is inactive"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate token
	token, err := h.jwtManager.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		log.Error().Err(err).Msg("Failed to generate token")
		return
	}

	// Update last login time
	if err := h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID); err != nil {
		// Log but don't fail the request
		log.Warn().Err(err).Int64("userID", user.ID).Msg("Failed to update last login time")
	}

	// Remove password hash from response
	user.PasswordHash = ""

	// Return token
	c.JSON(http.StatusOK, models.LoginResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: h.jwtManager.GetTokenExpiration(),
		User:      *user,
	})
}

// GetUsers handles requests to get all users
func (h *AuthHandler) GetUsers(c *gin.Context) {
	// Parse pagination parameters
	limit, offset := getPaginationParams(c)

	// Get users
	users, total, err := h.userRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		log.Error().Err(err).Msg("Failed to get users")
		return
	}

	// Remove password hashes
	for _, user := range users {
		user.PasswordHash = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUser handles requests to get a user by ID
func (h *AuthHandler) GetUser(c *gin.Context) {
	// Parse user ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user
	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		log.Error().Err(err).Int64("userID", id).Msg("Failed to get user")
		return
	}

	// Remove password hash
	user.PasswordHash = ""

	c.JSON(http.StatusOK, user)
}

// UpdateUser handles requests to update a user
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	// Parse user ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get the user to update
	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		log.Error().Err(err).Int64("userID", id).Msg("Failed to get user for update")
		return
	}

	// Parse request
	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user fields
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Password != "" {
		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			log.Error().Err(err).Msg("Failed to hash password")
			return
		}
		user.PasswordHash = string(hashedPassword)
	}
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Role != "" {
		user.Role = models.Role(req.Role)
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	user.UpdatedAt = time.Now()

	// Save updates
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		log.Error().Err(err).Int64("userID", id).Msg("Failed to update user")
		return
	}

	// Remove password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusOK, user)
}

// DeleteUser handles requests to delete a user
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	// Parse user ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Delete user
	if err := h.userRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		log.Error().Err(err).Int64("userID", id).Msg("Failed to delete user")
		return
	}

	c.Status(http.StatusNoContent)
}

// GetApiKeys handles requests to get API keys for a user
func (h *AuthHandler) GetApiKeys(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get API keys
	apiKeys, err := h.apiKeyManager.ListAPIKeys(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get API keys"})
		log.Error().Err(err).Int64("userID", userID.(int64)).Msg("Failed to get API keys")
		return
	}

	// Remove key value from response for security
	for _, key := range apiKeys {
		key.Key = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"api_keys": apiKeys,
		"count":    len(apiKeys),
	})
}

// CreateApiKey handles requests to create a new API key
func (h *AuthHandler) CreateApiKey(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse request
	var req models.APIKeyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create API key
	apiKey, err := h.apiKeyManager.GenerateAPIKey(&req, userID.(int64))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Error().Err(err).Int64("userID", userID.(int64)).Msg("Failed to create API key")
		return
	}

	c.JSON(http.StatusCreated, apiKey)
}

// DeleteApiKey handles requests to delete an API key
func (h *AuthHandler) DeleteApiKey(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse API key ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	// Delete API key
	if err := h.apiKeyManager.DeleteAPIKey(id, userID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		log.Error().Err(err).Int64("apiKeyID", id).Msg("Failed to delete API key")
		return
	}

	c.Status(http.StatusNoContent)
}
