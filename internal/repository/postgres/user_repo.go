package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/jackc/pgx/v5"
)

// UserRepository implements the repository.UserRepository interface
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *DB) repository.UserRepository {
	return &UserRepository{db: db}
}

// Create adds a new user to the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			username, email, password_hash, first_name, last_name,
			role, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $8
		) RETURNING id
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.FirstName,
		user.LastName, user.Role, user.IsActive, now,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name,
			role, is_active, last_login, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Role, &user.IsActive,
		&user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by their username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name,
			role, is_active, last_login, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Role, &user.IsActive,
		&user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by their email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name,
			role, is_active, last_login, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Role, &user.IsActive,
		&user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", email)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// List retrieves all users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM users`
	var total int
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with pagination
	query := `
		SELECT id, username, email, password_hash, first_name, last_name,
			role, is_active, last_login, created_at, updated_at
		FROM users
		ORDER BY username ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.FirstName, &user.LastName, &user.Role, &user.IsActive,
			&user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, total, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, first_name = $3, last_name = $4,
			role = $5, is_active = $6, updated_at = $7
		WHERE id = $8
	`

	now := time.Now()
	user.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.Role, user.IsActive, now, user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64) error {
	query := `UPDATE users SET last_login = $1 WHERE id = $2`
	now := time.Now()
	_, err := r.db.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// CreateAPIKey creates a new API key for a user
func (r *UserRepository) CreateAPIKey(ctx context.Context, apiKey *models.APIKey) error {
	query := `
		INSERT INTO api_keys (
			user_id, name, key, scope, expires_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id
	`

	now := time.Now()
	apiKey.CreatedAt = now

	err := r.db.QueryRow(ctx, query,
		apiKey.UserID, apiKey.Name, apiKey.Key, apiKey.Scope,
		apiKey.ExpiresAt, now,
	).Scan(&apiKey.ID)

	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	return nil
}

// GetAPIKeyByID retrieves an API key by its ID
func (r *UserRepository) GetAPIKeyByID(ctx context.Context, id int64) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key, scope, last_used, expires_at, created_at
		FROM api_keys
		WHERE id = $1
	`

	apiKey := &models.APIKey{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.Key,
		&apiKey.Scope, &apiKey.LastUsed, &apiKey.ExpiresAt, &apiKey.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("API key not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return apiKey, nil
}

// GetAPIKeyByKey retrieves an API key by its key value
func (r *UserRepository) GetAPIKeyByKey(ctx context.Context, key string) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key, scope, last_used, expires_at, created_at
		FROM api_keys
		WHERE key = $1
	`

	apiKey := &models.APIKey{}
	err := r.db.QueryRow(ctx, query, key).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.Key,
		&apiKey.Scope, &apiKey.LastUsed, &apiKey.ExpiresAt, &apiKey.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return apiKey, nil
}

// ListAPIKeys retrieves all API keys for a user
func (r *UserRepository) ListAPIKeys(ctx context.Context, userID int64) ([]*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key, scope, last_used, expires_at, created_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*models.APIKey
	for rows.Next() {
		apiKey := &models.APIKey{}
		if err := rows.Scan(
			&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.Key,
			&apiKey.Scope, &apiKey.LastUsed, &apiKey.ExpiresAt, &apiKey.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan API key row: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating API key rows: %w", err)
	}

	return apiKeys, nil
}

// DeleteAPIKey deletes an API key
func (r *UserRepository) DeleteAPIKey(ctx context.Context, id int64) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}
	return nil
}

// UpdateAPIKeyLastUsed updates the last used timestamp for an API key
func (r *UserRepository) UpdateAPIKeyLastUsed(ctx context.Context, id int64) error {
	query := `UPDATE api_keys SET last_used = $1 WHERE id = $2`
	now := time.Now()
	_, err := r.db.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to update API key last used: %w", err)
	}
	return nil
}
