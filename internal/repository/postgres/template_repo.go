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

// TemplateRepository implements the repository.TemplateRepository interface
type TemplateRepository struct {
	db *DB
}

// NewTemplateRepository creates a new TemplateRepository
func NewTemplateRepository(db *DB) repository.TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create adds a new template to the database
func (r *TemplateRepository) Create(ctx context.Context, template *models.Template) error {
	query := `
		INSERT INTO templates (
			name, description, max_length, sequence_start, sequence_length,
			sequence_padding, sequence_increment, sequence_position,
			created_by, created_at, updated_at, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, $11
		) RETURNING id
	`

	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		template.Name, template.Description, template.MaxLength,
		template.SequenceStart, template.SequenceLength, template.SequencePadding,
		template.SequenceIncrement, template.SequencePosition, template.CreatedBy,
		now, template.IsActive,
	).Scan(&template.ID)

	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

// GetByID retrieves a template by its ID
func (r *TemplateRepository) GetByID(ctx context.Context, id int64) (*models.Template, error) {
	query := `
		SELECT id, name, description, max_length, sequence_start, sequence_length,
			sequence_padding, sequence_increment, sequence_position,
			created_by, created_at, updated_at, is_active
		FROM templates
		WHERE id = $1
	`

	template := &models.Template{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&template.ID, &template.Name, &template.Description, &template.MaxLength,
		&template.SequenceStart, &template.SequenceLength, &template.SequencePadding,
		&template.SequenceIncrement, &template.SequencePosition, &template.CreatedBy,
		&template.CreatedAt, &template.UpdatedAt, &template.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("template not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Get template groups
	groups, err := r.GetTemplateGroups(ctx, template.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template groups: %w", err)
	}
	template.Groups = groups

	return template, nil
}

// GetByName retrieves a template by its name
func (r *TemplateRepository) GetByName(ctx context.Context, name string) (*models.Template, error) {
	query := `
		SELECT id, name, description, max_length, sequence_start, sequence_length,
			sequence_padding, sequence_increment, sequence_position,
			created_by, created_at, updated_at, is_active
		FROM templates
		WHERE name = $1
	`

	template := &models.Template{}
	err := r.db.QueryRow(ctx, query, name).Scan(
		&template.ID, &template.Name, &template.Description, &template.MaxLength,
		&template.SequenceStart, &template.SequenceLength, &template.SequencePadding,
		&template.SequenceIncrement, &template.SequencePosition, &template.CreatedBy,
		&template.CreatedAt, &template.UpdatedAt, &template.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("template not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Get template groups
	groups, err := r.GetTemplateGroups(ctx, template.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template groups: %w", err)
	}
	template.Groups = groups

	return template, nil
}

// List retrieves all templates with pagination
func (r *TemplateRepository) List(ctx context.Context, limit, offset int) ([]*models.Template, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM templates`
	var total int
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count templates: %w", err)
	}

	// Get templates with pagination
	query := `
		SELECT id, name, description, max_length, sequence_start, sequence_length,
			sequence_padding, sequence_increment, sequence_position,
			created_by, created_at, updated_at, is_active
		FROM templates
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.Template
	for rows.Next() {
		template := &models.Template{}
		if err := rows.Scan(
			&template.ID, &template.Name, &template.Description, &template.MaxLength,
			&template.SequenceStart, &template.SequenceLength, &template.SequencePadding,
			&template.SequenceIncrement, &template.SequencePosition, &template.CreatedBy,
			&template.CreatedAt, &template.UpdatedAt, &template.IsActive,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan template row: %w", err)
		}
		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating template rows: %w", err)
	}

	// Get groups for each template
	for _, template := range templates {
		groups, err := r.GetTemplateGroups(ctx, template.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get template groups: %w", err)
		}
		template.Groups = groups
	}

	return templates, total, nil
}

// Update updates an existing template
func (r *TemplateRepository) Update(ctx context.Context, template *models.Template) error {
	query := `
		UPDATE templates
		SET name = $1, description = $2, max_length = $3, sequence_start = $4,
			sequence_length = $5, sequence_padding = $6, sequence_increment = $7,
			sequence_position = $8, updated_at = $9, is_active = $10
		WHERE id = $11
	`

	now := time.Now()
	template.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		template.Name, template.Description, template.MaxLength,
		template.SequenceStart, template.SequenceLength, template.SequencePadding,
		template.SequenceIncrement, template.SequencePosition, now, template.IsActive,
		template.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	return nil
}

// Delete deletes a template
func (r *TemplateRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM templates WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	return nil
}

// GetTemplateGroups retrieves all groups for a template
func (r *TemplateRepository) GetTemplateGroups(ctx context.Context, templateID int64) ([]models.TemplateGroup, error) {
	query := `
		SELECT id, template_id, name, length, position, is_required, validation_type, validation_value
		FROM template_groups
		WHERE template_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.Query(ctx, query, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to query template groups: %w", err)
	}
	defer rows.Close()

	var groups []models.TemplateGroup
	for rows.Next() {
		var group models.TemplateGroup
		if err := rows.Scan(
			&group.ID, &group.TemplateID, &group.Name, &group.Length,
			&group.Position, &group.IsRequired, &group.ValidationType, &group.ValidationValue,
		); err != nil {
			return nil, fmt.Errorf("failed to scan template group row: %w", err)
		}
		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating template group rows: %w", err)
	}

	return groups, nil
}

// CreateTemplateGroup creates a new template group
func (r *TemplateRepository) CreateTemplateGroup(ctx context.Context, group *models.TemplateGroup) error {
	query := `
		INSERT INTO template_groups (
			template_id, name, length, position, is_required, validation_type, validation_value
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		group.TemplateID, group.Name, group.Length, group.Position,
		group.IsRequired, group.ValidationType, group.ValidationValue,
	).Scan(&group.ID)

	if err != nil {
		return fmt.Errorf("failed to create template group: %w", err)
	}

	return nil
}

// UpdateTemplateGroup updates an existing template group
func (r *TemplateRepository) UpdateTemplateGroup(ctx context.Context, group *models.TemplateGroup) error {
	query := `
		UPDATE template_groups
		SET name = $1, length = $2, position = $3, is_required = $4,
			validation_type = $5, validation_value = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(ctx, query,
		group.Name, group.Length, group.Position, group.IsRequired,
		group.ValidationType, group.ValidationValue, group.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update template group: %w", err)
	}

	return nil
}

// DeleteTemplateGroup deletes a template group
func (r *TemplateRepository) DeleteTemplateGroup(ctx context.Context, id int64) error {
	query := `DELETE FROM template_groups WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template group: %w", err)
	}
	return nil
}
