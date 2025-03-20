package repository

import (
	"context"

	"github.com/bilbothegreedy/HNS/internal/models"
)

// HostnameRepository defines the interface for hostname operations
type HostnameRepository interface {
	Create(ctx context.Context, hostname *models.Hostname) error
	GetByID(ctx context.Context, id int64) (*models.Hostname, error)
	GetByName(ctx context.Context, name string) (*models.Hostname, error)
	GetByStatus(ctx context.Context, status models.HostnameStatus, limit, offset int) ([]*models.Hostname, error)
	GetByTemplateID(ctx context.Context, templateID int64, limit, offset int) ([]*models.Hostname, error)
	UpdateStatus(ctx context.Context, id int64, status models.HostnameStatus, updatedBy string) error
	CommitHostname(ctx context.Context, id int64, committedBy string) error
	ReleaseHostname(ctx context.Context, id int64, releasedBy string) error
	GetNextSequenceNumber(ctx context.Context, templateID int64) (int, error)
	Count(ctx context.Context, templateID int64, status models.HostnameStatus) (int, error)
	List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*models.Hostname, int, error)
	CountByUser(ctx context.Context, username string, status models.HostnameStatus) (int, error)
}

// TemplateRepository defines the interface for template operations
type TemplateRepository interface {
	Create(ctx context.Context, template *models.Template) error
	GetByID(ctx context.Context, id int64) (*models.Template, error)
	GetByName(ctx context.Context, name string) (*models.Template, error)
	List(ctx context.Context, limit, offset int) ([]*models.Template, int, error)
	Update(ctx context.Context, template *models.Template) error
	Delete(ctx context.Context, id int64) error
	GetTemplateGroups(ctx context.Context, templateID int64) ([]models.TemplateGroup, error)
	CreateTemplateGroup(ctx context.Context, group *models.TemplateGroup) error
	UpdateTemplateGroup(ctx context.Context, group *models.TemplateGroup) error
	DeleteTemplateGroup(ctx context.Context, id int64) error
}

// UserRepository defines the interface for user operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, limit, offset int) ([]*models.User, int, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int64) error
	UpdateLastLogin(ctx context.Context, id int64) error

	// API Key operations
	CreateAPIKey(ctx context.Context, apiKey *models.APIKey) error
	GetAPIKeyByID(ctx context.Context, id int64) (*models.APIKey, error)
	GetAPIKeyByKey(ctx context.Context, key string) (*models.APIKey, error)
	ListAPIKeys(ctx context.Context, userID int64) ([]*models.APIKey, error)
	DeleteAPIKey(ctx context.Context, id int64) error
	UpdateAPIKeyLastUsed(ctx context.Context, id int64) error
}
