package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/jackc/pgx/v5"
)

// HostnameRepository implements the repository.HostnameRepository interface
type HostnameRepository struct {
	db *DB
}

// NewHostnameRepository creates a new HostnameRepository
func NewHostnameRepository(db *DB) repository.HostnameRepository {
	return &HostnameRepository{db: db}
}

// Create adds a new hostname to the database
func (r *HostnameRepository) Create(ctx context.Context, hostname *models.Hostname) error {
	query := `
		INSERT INTO hostnames (
			name, template_id, status, sequence_num, reserved_by, reserved_at,
			dns_verified, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $8
		) RETURNING id
	`

	now := time.Now()
	hostname.CreatedAt = now
	hostname.UpdatedAt = now
	hostname.ReservedAt = now

	err := r.db.QueryRow(ctx, query,
		hostname.Name, hostname.TemplateID, hostname.Status, hostname.SequenceNum,
		hostname.ReservedBy, hostname.ReservedAt, hostname.DNSVerified, now,
	).Scan(&hostname.ID)

	if err != nil {
		return fmt.Errorf("failed to create hostname: %w", err)
	}

	return nil
}

func (r *HostnameRepository) GetByID(ctx context.Context, id int64) (*models.Hostname, error) {
	query := `
		SELECT id, name, template_id, status, sequence_num, reserved_by, reserved_at,
			committed_by, committed_at, released_by, released_at, dns_verified,
			created_at, updated_at
		FROM hostnames
		WHERE id = $1
	`

	hostname := &models.Hostname{}

	// Temporary variables for handling NULL values
	var committedBy, releasedBy sql.NullString
	var committedAt, releasedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&hostname.ID, &hostname.Name, &hostname.TemplateID, &hostname.Status,
		&hostname.SequenceNum, &hostname.ReservedBy, &hostname.ReservedAt,
		&committedBy, &committedAt, &releasedBy,
		&releasedAt, &hostname.DNSVerified, &hostname.CreatedAt, &hostname.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("hostname not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Handle NULL value conversion
	if committedBy.Valid {
		hostname.CommittedBy = committedBy.String
	}
	if committedAt.Valid {
		hostname.CommittedAt = &committedAt.Time
	}
	if releasedBy.Valid {
		hostname.ReleasedBy = releasedBy.String
	}
	if releasedAt.Valid {
		hostname.ReleasedAt = &releasedAt.Time
	}

	return hostname, nil
}

// GetByName retrieves a hostname by its name
func (r *HostnameRepository) GetByName(ctx context.Context, name string) (*models.Hostname, error) {
	query := `
		SELECT id, name, template_id, status, sequence_num, reserved_by, reserved_at,
			committed_by, committed_at, released_by, released_at, dns_verified,
			created_at, updated_at
		FROM hostnames
		WHERE name = $1
	`

	hostname := &models.Hostname{}
	err := r.db.QueryRow(ctx, query, name).Scan(
		&hostname.ID, &hostname.Name, &hostname.TemplateID, &hostname.Status,
		&hostname.SequenceNum, &hostname.ReservedBy, &hostname.ReservedAt,
		&hostname.CommittedBy, &hostname.CommittedAt, &hostname.ReleasedBy,
		&hostname.ReleasedAt, &hostname.DNSVerified, &hostname.CreatedAt, &hostname.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("hostname not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	return hostname, nil
}

// GetByStatus retrieves hostnames by their status
func (r *HostnameRepository) GetByStatus(ctx context.Context, status models.HostnameStatus, limit, offset int) ([]*models.Hostname, error) {
	query := `
		SELECT id, name, template_id, status, sequence_num, reserved_by, reserved_at,
			committed_by, committed_at, released_by, released_at, dns_verified,
			created_at, updated_at
		FROM hostnames
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query hostnames: %w", err)
	}
	defer rows.Close()

	var hostnames []*models.Hostname
	for rows.Next() {
		hostname := &models.Hostname{}
		if err := rows.Scan(
			&hostname.ID, &hostname.Name, &hostname.TemplateID, &hostname.Status,
			&hostname.SequenceNum, &hostname.ReservedBy, &hostname.ReservedAt,
			&hostname.CommittedBy, &hostname.CommittedAt, &hostname.ReleasedBy,
			&hostname.ReleasedAt, &hostname.DNSVerified, &hostname.CreatedAt, &hostname.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan hostname row: %w", err)
		}
		hostnames = append(hostnames, hostname)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating hostname rows: %w", err)
	}

	return hostnames, nil
}

// GetByTemplateID retrieves hostnames by their template ID
func (r *HostnameRepository) GetByTemplateID(ctx context.Context, templateID int64, limit, offset int) ([]*models.Hostname, error) {
	query := `
		SELECT id, name, template_id, status, sequence_num, reserved_by, reserved_at,
			committed_by, committed_at, released_by, released_at, dns_verified,
			created_at, updated_at
		FROM hostnames
		WHERE template_id = $1
		ORDER BY sequence_num ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, templateID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query hostnames: %w", err)
	}
	defer rows.Close()

	var hostnames []*models.Hostname
	for rows.Next() {
		hostname := &models.Hostname{}
		if err := rows.Scan(
			&hostname.ID, &hostname.Name, &hostname.TemplateID, &hostname.Status,
			&hostname.SequenceNum, &hostname.ReservedBy, &hostname.ReservedAt,
			&hostname.CommittedBy, &hostname.CommittedAt, &hostname.ReleasedBy,
			&hostname.ReleasedAt, &hostname.DNSVerified, &hostname.CreatedAt, &hostname.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan hostname row: %w", err)
		}
		hostnames = append(hostnames, hostname)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating hostname rows: %w", err)
	}

	return hostnames, nil
}

// UpdateStatus updates the status of a hostname
func (r *HostnameRepository) UpdateStatus(ctx context.Context, id int64, status models.HostnameStatus, updatedBy string) error {
	query := `
		UPDATE hostnames
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, id, status, now)
	if err != nil {
		return fmt.Errorf("failed to update hostname status: %w", err)
	}

	return nil
}

// CommitHostname commits a reserved hostname
func (r *HostnameRepository) CommitHostname(ctx context.Context, id int64, committedBy string) error {
	query := `
		UPDATE hostnames
		SET status = $2, committed_by = $3, committed_at = $4, updated_at = $4
		WHERE id = $1 AND status = $5
	`

	now := time.Now()
	res, err := r.db.Exec(ctx, query, id, models.StatusCommitted, committedBy, now, models.StatusReserved)
	if err != nil {
		return fmt.Errorf("failed to commit hostname: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("hostname not found or not in reserved status")
	}

	return nil
}

// ReleaseHostname releases a committed hostname
func (r *HostnameRepository) ReleaseHostname(ctx context.Context, id int64, releasedBy string) error {
	query := `
		UPDATE hostnames
		SET status = $2, released_by = $3, released_at = $4, updated_at = $4
		WHERE id = $1 AND status = $5
	`

	now := time.Now()
	res, err := r.db.Exec(ctx, query, id, models.StatusReleased, releasedBy, now, models.StatusCommitted)
	if err != nil {
		return fmt.Errorf("failed to release hostname: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("hostname not found or not in committed status")
	}

	return nil
}

// GetNextSequenceNumber gets the next available sequence number for a template
func (r *HostnameRepository) GetNextSequenceNumber(ctx context.Context, templateID int64) (int, error) {
	query := `
		SELECT COALESCE(MAX(sequence_num), 0) + 1
		FROM hostnames
		WHERE template_id = $1
	`

	var nextSeq int
	err := r.db.QueryRow(ctx, query, templateID).Scan(&nextSeq)
	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence number: %w", err)
	}

	return nextSeq, nil
}

// Count counts hostnames by template ID and status
func (r *HostnameRepository) Count(ctx context.Context, templateID int64, status models.HostnameStatus) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM hostnames
		WHERE template_id = $1 AND status = $2
	`

	var count int
	err := r.db.QueryRow(ctx, query, templateID, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count hostnames: %w", err)
	}

	return count, nil
}

// List retrieves hostnames with filters
func (r *HostnameRepository) List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*models.Hostname, int, error) {
	// Base query
	query := `
		SELECT id, name, template_id, status, sequence_num, reserved_by, reserved_at,
			committed_by, committed_at, released_by, released_at, dns_verified,
			created_at, updated_at
		FROM hostnames
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM hostnames WHERE 1=1`

	// Build dynamic query parts
	args := []interface{}{}
	argCounter := 1
	whereClause := ""

	// Apply filters
	for key, value := range filters {
		whereClause += fmt.Sprintf(" AND %s = $%d", key, argCounter)
		args = append(args, value)
		argCounter++
	}

	// Add ordering and pagination
	query = query + whereClause + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	countQuery = countQuery + whereClause

	// Get total count first
	var total int
	err := r.db.QueryRow(ctx, countQuery, args[:argCounter-1]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count hostnames: %w", err)
	}

	// Query hostnames
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query hostnames: %w", err)
	}
	defer rows.Close()

	var hostnames []*models.Hostname
	for rows.Next() {
		hostname := &models.Hostname{}

		// Temporary variables for handling NULL values
		var committedBy, releasedBy sql.NullString
		var committedAt, releasedAt sql.NullTime

		if err := rows.Scan(
			&hostname.ID, &hostname.Name, &hostname.TemplateID, &hostname.Status,
			&hostname.SequenceNum, &hostname.ReservedBy, &hostname.ReservedAt,
			&committedBy, &committedAt, &releasedBy,
			&releasedAt, &hostname.DNSVerified, &hostname.CreatedAt, &hostname.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan hostname row: %w", err)
		}

		// Handle NULL value conversion
		if committedBy.Valid {
			hostname.CommittedBy = committedBy.String
		}
		if committedAt.Valid {
			hostname.CommittedAt = &committedAt.Time
		}
		if releasedBy.Valid {
			hostname.ReleasedBy = releasedBy.String
		}
		if releasedAt.Valid {
			hostname.ReleasedAt = &releasedAt.Time
		}

		hostnames = append(hostnames, hostname)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating hostname rows: %w", err)
	}

	return hostnames, total, nil
}
