package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// LoggerMiddleware logs incoming requests and their responses
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("requestID", requestID)

		// Get the request logger
		logger := utils.GetRequestLogger(requestID)

		// Set start time
		startTime := time.Now()

		// Get path without query parameters
		path := c.Request.URL.Path

		// Log request
		logger.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Msg("Request received")

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Log response
		logger.Info().
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Int("size", c.Writer.Size()).
			Int("error_count", len(c.Errors)).
			Msg("Request completed")
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

// JWTAuthMiddleware validates JWT tokens in the Authorization header
func JWTAuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate the token
		claims, err := jwtManager.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// APIKeyAuthMiddleware validates API keys in the X-API-Key header
func APIKeyAuthMiddleware(apiKeyManager *auth.APIKeyManager, requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the API key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return
		}

		// Validate the API key
		key, err := apiKeyManager.ValidateAPIKey(apiKey, requiredScope)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired API key"})
			c.Abort()
			return
		}

		// Set API key information in context
		c.Set("apiKeyID", key.ID)
		c.Set("apiKeyUserID", key.UserID)
		c.Set("apiKeyScope", key.Scope)

		c.Next()
	}
}

// AuthMiddleware is a combined authentication middleware that supports both JWT and API key
func AuthMiddleware(jwtManager *auth.JWTManager, apiKeyManager *auth.APIKeyManager, requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try API key first
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			// Validate the API key
			key, err := apiKeyManager.ValidateAPIKey(apiKey, requiredScope)
			if err == nil {
				// API key is valid
				c.Set("apiKeyID", key.ID)
				c.Set("apiKeyUserID", key.UserID)
				c.Set("apiKeyScope", key.Scope)
				c.Set("authMethod", "apikey")
				c.Next()
				return
			}
		}

		// Try JWT token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Check if it's a Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]

				// Validate the token
				claims, err := jwtManager.VerifyToken(tokenString)
				if err == nil {
					// Token is valid
					c.Set("userID", claims.UserID)
					c.Set("username", claims.Username)
					c.Set("email", claims.Email)
					c.Set("role", claims.Role)
					c.Set("authMethod", "jwt")
					c.Next()
					return
				}
			}
		}

		// Neither API key nor JWT token is valid
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		c.Abort()
	}
}

// RoleMiddleware checks if the authenticated user has the required role
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user role from context
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role information not available"})
			c.Abort()
			return
		}

		// Check if the role matches the required role
		if role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
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
				// Get request ID if available
				requestID, exists := c.Get("requestID")
				var logger = log.Logger
				if exists {
					logger = utils.GetRequestLogger(requestID.(string))
				}

				// Log the error
				logger.Error().
					Interface("error", err).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Msg("Panic recovered")

				// Return error response
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()

		c.Next()
	}
}
