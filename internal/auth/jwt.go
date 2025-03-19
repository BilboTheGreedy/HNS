package auth

import (
	"fmt"
	"time"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

// JWTManager is responsible for JWT token operations
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWTManager
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken generates a new JWT token for a user
func (m *JWTManager) GenerateToken(user *models.User) (string, error) {
	now := time.Now()
	expirationTime := now.Add(m.tokenDuration)

	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "HNS",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to create signed token: %w", err)
	}

	return signedToken, nil
}

// VerifyToken verifies the JWT token and returns the claims
func (m *JWTManager) VerifyToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GetTokenExpiration returns the token expiration in seconds
func (m *JWTManager) GetTokenExpiration() int64 {
	return int64(m.tokenDuration.Seconds())
}
