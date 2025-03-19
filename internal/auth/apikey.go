package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
)

// APIKeyManager manages API keys
type APIKeyManager struct {
	userRepo      repository.UserRepository
	keyExpiration time.Duration
}

// NewAPIKeyManager creates a new APIKeyManager
func NewAPIKeyManager(userRepo repository.UserRepository, keyExpiration time.Duration) *APIKeyManager {
	return &APIKeyManager{
		userRepo:      userRepo,
		keyExpiration: keyExpiration,
	}
}

// GenerateAPIKey generates a new API key for a user
func (m *APIKeyManager) GenerateAPIKey(req *models.APIKeyCreateRequest, userID int64) (*models.APIKeyResponse, error) {
	// Validate scope
	if err := m.validateScope(req.Scope); err != nil {
		return nil, err
	}

	// Generate a random API key
	key, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Create the API key record
	apiKey := &models.APIKey{
		UserID:    userID,
		Name:      req.Name,
		Key:       key,
		Scope:     req.Scope,
		ExpiresAt: time.Now().Add(m.keyExpiration),
	}

	// Save the API key
	if err := m.userRepo.CreateAPIKey(context.Background(), apiKey); err != nil {
		return nil, fmt.Errorf("failed to save API key: %w", err)
	}

	// Create the response
	response := &models.APIKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       apiKey.Key,
		Scope:     apiKey.Scope,
		ExpiresAt: apiKey.ExpiresAt,
	}

	return response, nil
}

// ValidateAPIKey validates an API key and checks its scope
func (m *APIKeyManager) ValidateAPIKey(key string, requiredScope string) (*models.APIKey, error) {
	// Get the API key
	apiKey, err := m.userRepo.GetAPIKeyByKey(context.Background(), key)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check if the key has expired
	if apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Check if the key has the required scope
	if !m.hasRequiredScope(apiKey.Scope, requiredScope) {
		return nil, fmt.Errorf("API key does not have the required scope")
	}

	// Update last used timestamp
	if err := m.userRepo.UpdateAPIKeyLastUsed(context.Background(), apiKey.ID); err != nil {
		// Log but don't fail the request
		fmt.Printf("Failed to update API key last used timestamp: %v\n", err)
	}

	return apiKey, nil
}

// ListAPIKeys lists API keys for a user
func (m *APIKeyManager) ListAPIKeys(userID int64) ([]*models.APIKey, error) {
	return m.userRepo.ListAPIKeys(context.Background(), userID)
}

// DeleteAPIKey deletes an API key
func (m *APIKeyManager) DeleteAPIKey(keyID int64, userID int64) error {
	// Get the API key
	apiKey, err := m.userRepo.GetAPIKeyByID(context.Background(), keyID)
	if err != nil {
		return fmt.Errorf("API key not found")
	}

	// Check if the key belongs to the user
	if apiKey.UserID != userID {
		return fmt.Errorf("API key does not belong to the user")
	}

	// Delete the API key
	if err := m.userRepo.DeleteAPIKey(context.Background(), keyID); err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// validateScope validates the scope format
func (m *APIKeyManager) validateScope(scope string) error {
	// Validate scope format (comma-separated list of allowed scopes)
	allowedScopes := []string{"read", "reserve", "commit", "release", "admin"}
	scopes := strings.Split(scope, ",")

	for _, s := range scopes {
		s = strings.TrimSpace(s)
		isValid := false
		for _, allowed := range allowedScopes {
			if s == allowed {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid scope: %s", s)
		}
	}

	return nil
}

// hasRequiredScope checks if the API key scope includes the required scope
func (m *APIKeyManager) hasRequiredScope(keyScope, requiredScope string) bool {
	// If the key has admin scope, it has all scopes
	if strings.Contains(keyScope, "admin") {
		return true
	}

	// Split the key scope into individual scopes
	scopes := strings.Split(keyScope, ",")
	for _, scope := range scopes {
		if strings.TrimSpace(scope) == requiredScope {
			return true
		}
	}

	return false
}
