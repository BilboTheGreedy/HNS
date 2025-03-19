package models

import (
	"time"
)

// Role represents a user role
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// User represents a user in the system
type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Role         Role      `json:"role" db:"role"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	LastLogin    *time.Time `json:"last_login,omitempty" db:"last_login"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserCreateRequest represents a request to create a new user
type UserCreateRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Role      string `json:"role" binding:"required,oneof=admin user"`
}

// UserUpdateRequest represents a request to update an existing user
type UserUpdateRequest struct {
	Email     string `json:"email" binding:"omitempty,email"`
	Password  string `json:"password" binding:"omitempty,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role" binding:"omitempty,oneof=admin user"`
	IsActive  *bool  `json:"is_active"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a response to a login request
type LoginResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn int64  `json:"expires_in"` // in seconds
	User      User   `json:"user"`
}

// APIKey represents an API key
type APIKey struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Key         string    `json:"key,omitempty" db:"key"`
	Scope       string    `json:"scope" db:"scope"`
	LastUsed    *time.Time `json:"last_used,omitempty" db:"last_used"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// APIKeyCreateRequest represents a request to create a new API key
type APIKeyCreateRequest struct {
	Name  string `json:"name" binding:"required"`
	Scope string `json:"scope" binding:"required"`
}

// APIKeyResponse represents a response with the API key information
type APIKeyResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
}