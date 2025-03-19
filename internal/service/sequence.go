package service

import (
	"context"
	"fmt"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
)

// SequenceService is responsible for managing sequence numbers
type SequenceService struct {
	hostnameRepo repository.HostnameRepository
}

// NewSequenceService creates a new SequenceService
func NewSequenceService(hostnameRepo repository.HostnameRepository) *SequenceService {
	return &SequenceService{
		hostnameRepo: hostnameRepo,
	}
}

// GetNextSequenceNumber returns the next available sequence number for a template
func (s *SequenceService) GetNextSequenceNumber(ctx context.Context, templateID int64) (int, error) {
	return s.hostnameRepo.GetNextSequenceNumber(ctx, templateID)
}

// ReserveSequenceNumber reserves a specific sequence number for a template
func (s *SequenceService) ReserveSequenceNumber(ctx context.Context, templateID int64, sequenceNum int, reservedBy string) error {
	// Check if the sequence number is already in use
	hostnames, err := s.hostnameRepo.GetByTemplateID(ctx, templateID, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to check existing hostnames: %w", err)
	}

	for _, hostname := range hostnames {
		if hostname.SequenceNum == sequenceNum {
			return fmt.Errorf("sequence number %d is already in use for template %d", sequenceNum, templateID)
		}
	}

	// Create a placeholder hostname to reserve the sequence
	placeholder := &models.Hostname{
		Name:        fmt.Sprintf("RESERVED-SEQ-%d", sequenceNum),
		TemplateID:  templateID,
		Status:      models.StatusReserved,
		SequenceNum: sequenceNum,
		ReservedBy:  reservedBy,
		DNSVerified: false,
	}

	// Save to database
	if err := s.hostnameRepo.Create(ctx, placeholder); err != nil {
		return fmt.Errorf("failed to reserve sequence number: %w", err)
	}

	return nil
}

// ReleaseSequenceNumber releases a reserved sequence number
func (s *SequenceService) ReleaseSequenceNumber(ctx context.Context, templateID int64, sequenceNum int) error {
	// Find the placeholder hostname
	filters := map[string]interface{}{
		"template_id":  templateID,
		"sequence_num": sequenceNum,
	}
	hostnames, _, err := s.hostnameRepo.List(ctx, 1, 0, filters)
	if err != nil || len(hostnames) == 0 {
		return fmt.Errorf("failed to find reserved sequence number: %w", err)
	}

	// Update status to released
	if err := s.hostnameRepo.UpdateStatus(ctx, hostnames[0].ID, models.StatusReleased, "system"); err != nil {
		return fmt.Errorf("failed to release sequence number: %w", err)
	}

	return nil
}

// GetSequenceUsage returns information about sequence number usage for a template
type SequenceUsageInfo struct {
	TemplateID      int64 `json:"template_id"`
	TotalSequences  int   `json:"total_sequences"`
	UsedSequences   int   `json:"used_sequences"`
	NextSequence    int   `json:"next_sequence"`
	HighestSequence int   `json:"highest_sequence"`
	LowestSequence  int   `json:"lowest_sequence"`
}

// GetSequenceUsage returns information about sequence number usage for a template
func (s *SequenceService) GetSequenceUsage(ctx context.Context, templateID int64) (*SequenceUsageInfo, error) {
	// Get all hostnames for the template
	hostnames, err := s.hostnameRepo.GetByTemplateID(ctx, templateID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get hostnames: %w", err)
	}

	// Count and find min/max
	usage := &SequenceUsageInfo{
		TemplateID:      templateID,
		TotalSequences:  len(hostnames),
		UsedSequences:   0,
		NextSequence:    0,
		HighestSequence: 0,
		LowestSequence:  0,
	}

	// Calculate statistics
	if len(hostnames) > 0 {
		usage.LowestSequence = hostnames[0].SequenceNum
		usage.HighestSequence = hostnames[0].SequenceNum

		for _, hostname := range hostnames {
			// Count used sequences (reserved or committed)
			if hostname.Status == models.StatusReserved || hostname.Status == models.StatusCommitted {
				usage.UsedSequences++
			}

			// Update highest/lowest
			if hostname.SequenceNum > usage.HighestSequence {
				usage.HighestSequence = hostname.SequenceNum
			}
			if hostname.SequenceNum < usage.LowestSequence {
				usage.LowestSequence = hostname.SequenceNum
			}
		}

		// Get next sequence
		usage.NextSequence, err = s.hostnameRepo.GetNextSequenceNumber(ctx, templateID)
		if err != nil {
			return nil, fmt.Errorf("failed to get next sequence number: %w", err)
		}
	}

	return usage, nil
}

// FindSequenceGaps finds gaps in the sequence numbers for a template
func (s *SequenceService) FindSequenceGaps(ctx context.Context, templateID int64, maxGaps int) ([]int, error) {
	// Get all hostnames for the template
	hostnames, err := s.hostnameRepo.GetByTemplateID(ctx, templateID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get hostnames: %w", err)
	}

	// Create a map of used sequence numbers
	usedSequences := make(map[int]bool)
	var max, min int
	if len(hostnames) > 0 {
		min = hostnames[0].SequenceNum
		max = hostnames[0].SequenceNum

		for _, hostname := range hostnames {
			usedSequences[hostname.SequenceNum] = true
			if hostname.SequenceNum > max {
				max = hostname.SequenceNum
			}
			if hostname.SequenceNum < min {
				min = hostname.SequenceNum
			}
		}
	} else {
		// No hostnames yet
		return []int{}, nil
	}

	// Find gaps
	var gaps []int
	for i := min; i <= max && len(gaps) < maxGaps; i++ {
		if !usedSequences[i] {
			gaps = append(gaps, i)
		}
	}

	return gaps, nil
}
