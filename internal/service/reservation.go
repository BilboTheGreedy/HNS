package service

import (
	"context"
	"fmt"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/rs/zerolog/log"
)

// ReservationService is responsible for hostname reservation operations
type ReservationService struct {
	hostnameRepo repository.HostnameRepository
	templateRepo repository.TemplateRepository
	generatorSvc *GeneratorService
}

// NewReservationService creates a new ReservationService
func NewReservationService(hostnameRepo repository.HostnameRepository, templateRepo repository.TemplateRepository) *ReservationService {
	return &ReservationService{
		hostnameRepo: hostnameRepo,
		templateRepo: templateRepo,
		generatorSvc: NewGeneratorService(templateRepo),
	}
}

// ReserveHostname reserves a hostname based on template and parameters
func (s *ReservationService) ReserveHostname(ctx context.Context, req *models.HostnameReservationRequest) (*models.Hostname, error) {
	// Get template
	template, err := s.templateRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Get next sequence number
	nextSeq, err := s.hostnameRepo.GetNextSequenceNumber(ctx, req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next sequence number: %w", err)
	}

	// Generate hostname
	hostnameStr, err := s.generatorSvc.GenerateHostname(ctx, req.TemplateID, nextSeq, req.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hostname: %w", err)
	}

	// Check if hostname already exists
	existing, err := s.hostnameRepo.GetByName(ctx, hostnameStr)
	if err == nil && existing != nil {
		// Hostname already exists, try with incremented sequence number
		log.Info().
			Str("hostname", hostnameStr).
			Int("sequence", nextSeq).
			Msg("Hostname already exists, trying with incremented sequence")

		// Increment by the template's increment value
		nextSeq += template.SequenceIncrement
		hostnameStr, err = s.generatorSvc.GenerateHostname(ctx, req.TemplateID, nextSeq, req.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to generate hostname with incremented sequence: %w", err)
		}

		// Check again to make sure this one is available
		existing, err = s.hostnameRepo.GetByName(ctx, hostnameStr)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("generated hostname still exists after incrementing sequence")
		}
	}

	// Create hostname record
	hostname := &models.Hostname{
		Name:        hostnameStr,
		TemplateID:  req.TemplateID,
		Status:      models.StatusReserved,
		SequenceNum: nextSeq,
		ReservedBy:  req.RequestedBy,
		DNSVerified: false,
	}

	// Save to database
	if err := s.hostnameRepo.Create(ctx, hostname); err != nil {
		return nil, fmt.Errorf("failed to create hostname record: %w", err)
	}

	return hostname, nil
}

// CommitHostname commits a reserved hostname
func (s *ReservationService) CommitHostname(ctx context.Context, req *models.HostnameCommitRequest) error {
	// Check if hostname exists and is reserved
	hostname, err := s.hostnameRepo.GetByID(ctx, req.HostnameID)
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	if hostname.Status != models.StatusReserved {
		return fmt.Errorf("hostname is not in reserved status, current status: %s", hostname.Status)
	}

	// Commit the hostname
	if err := s.hostnameRepo.CommitHostname(ctx, req.HostnameID, req.CommittedBy); err != nil {
		return fmt.Errorf("failed to commit hostname: %w", err)
	}

	return nil
}

// ReleaseHostname releases a committed hostname
func (s *ReservationService) ReleaseHostname(ctx context.Context, req *models.HostnameReleaseRequest) error {
	// Check if hostname exists and is committed
	hostname, err := s.hostnameRepo.GetByID(ctx, req.HostnameID)
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	if hostname.Status != models.StatusCommitted {
		return fmt.Errorf("hostname is not in committed status, current status: %s", hostname.Status)
	}

	// Release the hostname
	if err := s.hostnameRepo.ReleaseHostname(ctx, req.HostnameID, req.ReleasedBy); err != nil {
		return fmt.Errorf("failed to release hostname: %w", err)
	}

	return nil
}

// GetReservedHostnames gets all reserved hostnames
func (s *ReservationService) GetReservedHostnames(ctx context.Context, limit, offset int) ([]*models.Hostname, error) {
	return s.hostnameRepo.GetByStatus(ctx, models.StatusReserved, limit, offset)
}

// GetCommittedHostnames gets all committed hostnames
func (s *ReservationService) GetCommittedHostnames(ctx context.Context, limit, offset int) ([]*models.Hostname, error) {
	return s.hostnameRepo.GetByStatus(ctx, models.StatusCommitted, limit, offset)
}

// GetHostnamesByTemplate gets all hostnames for a template
func (s *ReservationService) GetHostnamesByTemplate(ctx context.Context, templateID int64, limit, offset int) ([]*models.Hostname, error) {
	return s.hostnameRepo.GetByTemplateID(ctx, templateID, limit, offset)
}

// GetHostname gets a hostname by ID
func (s *ReservationService) GetHostname(ctx context.Context, id int64) (*models.Hostname, error) {
	return s.hostnameRepo.GetByID(ctx, id)
}

// SearchHostnames searches for hostnames using filters
func (s *ReservationService) SearchHostnames(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Hostname, int, error) {
	return s.hostnameRepo.List(ctx, limit, offset, filters)
}
