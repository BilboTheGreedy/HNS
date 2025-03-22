package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// APIKeyHandler handles API key-related requests
type APIKeyHandler struct {
	BaseHandler
	userRepo      repository.UserRepository
	apiKeyManager *auth.APIKeyManager
}

// NewAPIKeyHandler creates a new APIKeyHandler
func NewAPIKeyHandler(userRepo repository.UserRepository, apiKeyManager *auth.APIKeyManager) *APIKeyHandler {
	return &APIKeyHandler{
		BaseHandler:   BaseHandler{},
		userRepo:      userRepo,
		apiKeyManager: apiKeyManager,
	}
}

// APIKeysList displays the user's API keys
func (h *APIKeyHandler) APIKeysList(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		h.RedirectWithAlert(c, "/login", "danger", "User not authenticated")
		return
	}

	// Get API keys
	ctx := c.Request.Context()
	apiKeys, err := h.userRepo.ListAPIKeys(ctx, userID.(int64))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get API keys")
		h.RenderTemplate(c, "apikey_list", gin.H{
			"Title":   "API Keys",
			"ApiKeys": []models.APIKey{},
		})
		return
	}

	// Render template
	h.RenderTemplate(c, "apikey_list", gin.H{
		"Title":      "API Keys",
		"ActivePage": "api-keys",
		"ApiKeys":    apiKeys,
		"Scopes":     []string{"read", "reserve", "commit", "release", "admin"}, // Available scopes
	})
}

// CreateAPIKey handles API key creation
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		h.RedirectWithAlert(c, "/login", "danger", "User not authenticated")
		return
	}

	// Get form data
	name := c.PostForm("name")
	if name == "" {
		h.RedirectWithAlert(c, "/api-keys", "danger", "API key name is required")
		return
	}

	// Get scope (multiple checkboxes)
	scopeArray := c.PostFormArray("scopes")
	if len(scopeArray) == 0 {
		h.RedirectWithAlert(c, "/api-keys", "danger", "At least one scope must be selected")
		return
	}
	scope := strings.Join(scopeArray, ",")

	// Create API key
	req := &models.APIKeyCreateRequest{
		Name:  name,
		Scope: scope,
	}

	// Use API key manager to create the key
	apiKey, err := h.apiKeyManager.GenerateAPIKey(req, userID.(int64))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create API key")
		h.RedirectWithAlert(c, "/api-keys", "danger", "Failed to create API key: "+err.Error())
		return
	}

	// Render success page with the key details
	// We need to show the key here because it's only visible once
	h.RenderTemplate(c, "apikey_created", gin.H{
		"Title":      "API Key Created",
		"ActivePage": "api-keys",
		"APIKey":     apiKey,
	})
}

// DeleteAPIKey handles API key deletion
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		h.RedirectWithAlert(c, "/login", "danger", "User not authenticated")
		return
	}

	// Get API key ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/api-keys", "danger", "Invalid API key ID")
		return
	}

	// Delete API key
	err = h.apiKeyManager.DeleteAPIKey(id, userID.(int64))
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete API key")
		h.RedirectWithAlert(c, "/api-keys", "danger", "Failed to delete API key: "+err.Error())
		return
	}

	// Redirect with success message
	h.RedirectWithAlert(c, "/api-keys", "success", "API key deleted successfully")
}

// CreateAPIKeyManager creates an API key manager if needed
func CreateAPIKeyManager(userRepo repository.UserRepository) *auth.APIKeyManager {
	// Default to 30 days expiration
	return auth.NewAPIKeyManager(userRepo, 30*24*time.Hour)
}
