// internal/auth/auth_handler.go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related operations
type AuthHandler struct {
	userRepo      repository.UserRepository
	jwtManager    *JWTManager
	apiKeyManager *APIKeyManager
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo repository.UserRepository, jwtManager *JWTManager, apiKeyManager *APIKeyManager) *AuthHandler {
	return &AuthHandler{
		userRepo:      userRepo,
		jwtManager:    jwtManager,
		apiKeyManager: apiKeyManager,
	}
}

// RegisterUser handles user registration requests
func (h *AuthHandler) RegisterUser(ctx context.Context, req *models.UserCreateRequest) error {
	// Check if username already exists
	existingUser, err := h.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return fmt.Errorf("username already exists")
	}

	// Check if email already exists
	existingUser, err = h.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
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
	if err := h.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// ValidateCredentials validates user credentials and returns a JWT token
func (h *AuthHandler) ValidateCredentials(ctx context.Context, req *models.LoginRequest) (string, error) {
	// Get user by username
	user, err := h.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return "", fmt.Errorf("invalid username or password")
	}

	// Check if user is active
	if !user.IsActive {
		return "", fmt.Errorf("user account is inactive")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", fmt.Errorf("invalid username or password")
	}

	// Generate token
	token, err := h.jwtManager.GenerateToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login time
	if err := h.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log but don't fail the request
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	return token, nil
}

// RegisterUserInternal is a version of RegisterUser that returns the created user
func (h *AuthHandler) RegisterUserInternal(ctx context.Context, req *models.UserCreateRequest) (*models.User, error) {
	// Check if username already exists
	existingUser, err := h.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	existingUser, err = h.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
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
	if err := h.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
