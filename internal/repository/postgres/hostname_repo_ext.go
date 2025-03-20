package postgres

import (
	"context"
	"fmt"

	"github.com/bilbothegreedy/HNS/internal/models"
)

// CountByUser counts hostnames by user and status
func (r *HostnameRepository) CountByUser(ctx context.Context, username string, status models.HostnameStatus) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM hostnames
		WHERE reserved_by = $1 AND status = $2
	`

	var count int
	err := r.db.QueryRow(ctx, query, username, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count hostnames by user: %w", err)
	}

	return count, nil
}

// ListByUser retrieves hostnames by a specific user
func (r *HostnameRepository) ListByUser(ctx context.Context, username string, limit, offset int) ([]*models.Hostname, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM hostnames WHERE reserved_by = $1`
	var total int
	err := r.db.QueryRow(ctx, countQuery, username).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count hostnames by user: %w", err)
	}

	// Get hostnames
	query := `
		SELECT id, name, template_id, status, sequence_num, reserved_by, reserved_at,
			committed_by, committed_at, released_by, released_at, dns_verified,
			created_at, updated_at
		FROM hostnames
		WHERE reserved_by = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, username, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query hostnames: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan hostname row: %w", err)
		}
		hostnames = append(hostnames, hostname)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating hostname rows: %w", err)
	}

	return hostnames, total, nil
}


package postgres

import (
	"context"
	"fmt"

	"github.com/bilbothegreedy/HNS/internal/models"
)

// CountByUser counts hostnames by user and status
func (r *HostnameRepository) CountByUser(ctx context.Context, username string, status models.HostnameStatus) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM hostnames
		WHERE reserved_by = $1 AND status = $2
	`

	var count int
	err := r.db.QueryRow(ctx, query, username, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count hostnames by user: %w", err)
	}

	return count, nil
}

// ListByUser retrieves hostnames by a specific user
func (r *HostnameRepository) ListByUser(ctx context.Context, username string, limit, offset int) ([]*models.Hostname, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM hostnames WHERE reserved_by = $1`
	var total int
	err := r.db.QueryRow(ctx, countQuery, username).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count hostnames by user: %w", err)
	}

	// Get hostnames
	query := `
		SELECT id, name, template_id, status, sequence_num, reserved_by, reserved_at,
			committed_by, committed_at, released_by, released_at, dns_verified,
			created_at, updated_at
		FROM hostnames
		WHERE reserved_by = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, username, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query hostnames: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan hostname row: %w", err)
		}
		hostnames = append(hostnames, hostname)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating hostname rows: %w", err)
	}

	return hostnames, total, nil
}