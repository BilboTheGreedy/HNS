package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/rs/zerolog/log"
)

// GeneratorService is responsible for generating hostnames
type GeneratorService struct {
	templateRepo repository.TemplateRepository
}

// NewGeneratorService creates a new GeneratorService
func NewGeneratorService(templateRepo repository.TemplateRepository) *GeneratorService {
	return &GeneratorService{
		templateRepo: templateRepo,
	}
}

// GenerateHostname generates a hostname based on a template and parameters
func (s *GeneratorService) GenerateHostname(ctx context.Context, templateID int64, sequenceNum int, params map[string]string) (string, error) {
	// Fetch the template and its groups
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return "", fmt.Errorf("failed to get template: %w", err)
	}

	// If the sequence number is not provided, use the template's start sequence
	if sequenceNum <= 0 {
		sequenceNum = template.SequenceStart
	}

	// Format the sequence based on template settings
	sequenceStr := formatSequence(sequenceNum, template.SequenceLength, template.SequencePadding)

	// Generate the hostname
	hostname := s.buildHostname(template, sequenceStr, params)

	// Validate hostname length
	if len(hostname) > template.MaxLength {
		return "", fmt.Errorf("generated hostname exceeds maximum length of %d characters", template.MaxLength)
	}

	return hostname, nil
}

// formatSequence formats a sequence number based on length and padding
func formatSequence(num, length int, padding bool) string {
	if padding {
		return fmt.Sprintf("%0*d", length, num)
	}
	return strconv.Itoa(num)
}

// buildHostname builds a hostname from a template and parameters
func (s *GeneratorService) buildHostname(template *models.Template, sequenceStr string, params map[string]string) string {
	// Groups to build the hostname
	var hostnameBuilder strings.Builder

	// Sort groups by position
	groupsByPosition := make(map[int]models.TemplateGroup)
	for _, group := range template.Groups {
		groupsByPosition[group.Position] = group
	}

	// Build the hostname by processing each group
	for i := 1; i <= len(template.Groups); i++ {
		group, exists := groupsByPosition[i]
		if !exists {
			continue
		}

		// Handle different validation types
		var groupValue string
		switch group.ValidationType {
		case string(models.ValidationTypeFixed):
			// Fixed value from the validation value
			groupValue = group.ValidationValue
		case string(models.ValidationTypeSequence):
			// Use the sequence number
			groupValue = sequenceStr
		default:
			// Try to get value from parameters
			paramValue, exists := params[group.Name]
			if exists {
				// Validate against the validation rule if provided
				if group.ValidationType == string(models.ValidationTypeRegex) && group.ValidationValue != "" {
					if match, _ := regexp.MatchString(group.ValidationValue, paramValue); !match {
						log.Warn().
							Str("group", group.Name).
							Str("value", paramValue).
							Str("pattern", group.ValidationValue).
							Msg("Group value does not match validation pattern")

						// If required, use a default or empty value
						if group.IsRequired {
							// Use first character of validation value as default if possible
							if len(group.ValidationValue) > 0 {
								groupValue = string(group.ValidationValue[0])
							} else {
								groupValue = "X" // Fallback default
							}
						} else {
							groupValue = "" // Optional group, skip it
						}
					} else {
						groupValue = paramValue
					}
				} else if group.ValidationType == string(models.ValidationTypeList) && group.ValidationValue != "" {
					// Check if value is in the allowed list
					allowedValues := strings.Split(group.ValidationValue, ",")
					allowed := false
					for _, allowedValue := range allowedValues {
						if strings.TrimSpace(allowedValue) == paramValue {
							allowed = true
							break
						}
					}

					if !allowed {
						log.Warn().
							Str("group", group.Name).
							Str("value", paramValue).
							Str("allowed", group.ValidationValue).
							Msg("Group value not in allowed list")

						// If required, use the first value from the list, otherwise skip
						if group.IsRequired && len(allowedValues) > 0 {
							groupValue = strings.TrimSpace(allowedValues[0])
						} else {
							groupValue = ""
						}
					} else {
						groupValue = paramValue
					}
				} else {
					// No validation or validation passed
					groupValue = paramValue
				}
			} else if group.IsRequired {
				// Required parameter not provided, use default or error
				log.Warn().
					Str("group", group.Name).
					Msg("Required group parameter not provided")

				// Use a default value
				if group.ValidationType == string(models.ValidationTypeList) && group.ValidationValue != "" {
					// Use first value from list
					allowedValues := strings.Split(group.ValidationValue, ",")
					if len(allowedValues) > 0 {
						groupValue = strings.TrimSpace(allowedValues[0])
					} else {
						groupValue = "X" // Fallback default
					}
				} else {
					groupValue = "X" // Fallback default
				}
			}
		}

		// Trim to maximum length if specified
		if group.Length > 0 && len(groupValue) > group.Length {
			groupValue = groupValue[:group.Length]
		}

		// Add to hostname
		hostnameBuilder.WriteString(groupValue)
	}

	return hostnameBuilder.String()
}

// ValidateTemplate validates a template definition
func (s *GeneratorService) ValidateTemplate(ctx context.Context, template *models.Template) error {
	// Check basic requirements
	if template.MaxLength <= 0 {
		return fmt.Errorf("template max length must be positive")
	}

	if template.SequenceLength <= 0 {
		return fmt.Errorf("sequence length must be positive")
	}

	if template.SequenceIncrement <= 0 {
		return fmt.Errorf("sequence increment must be positive")
	}

	// Check that groups don't exceed max length
	totalLength := 0
	for _, group := range template.Groups {
		totalLength += group.Length
	}

	if totalLength > template.MaxLength {
		return fmt.Errorf("sum of group lengths (%d) exceeds template max length (%d)", totalLength, template.MaxLength)
	}

	return nil
}

// GetAvailableTemplates returns all available templates
func (s *GeneratorService) GetAvailableTemplates(ctx context.Context, limit, offset int) ([]*models.Template, int, error) {
	return s.templateRepo.List(ctx, limit, offset)
}

// GetTemplateByID returns a template by ID
func (s *GeneratorService) GetTemplateByID(ctx context.Context, id int64) (*models.Template, error) {
	return s.templateRepo.GetByID(ctx, id)
}

// CreateTemplate creates a new template
func (s *GeneratorService) CreateTemplate(ctx context.Context, req *models.TemplateCreateRequest) (*models.Template, error) {
	// Create template object
	template := &models.Template{
		Name:              req.Name,
		Description:       req.Description,
		MaxLength:         req.MaxLength,
		SequenceStart:     req.SequenceStart,
		SequenceLength:    req.SequenceLength,
		SequencePadding:   req.SequencePadding,
		SequenceIncrement: req.SequenceIncrement,
		CreatedBy:         req.CreatedBy,
		IsActive:          true,
	}

	// Validate template
	if err := s.ValidateTemplate(ctx, template); err != nil {
		return nil, err
	}

	// Save template
	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	// Process and save groups
	for i, groupReq := range req.Groups {
		group := &models.TemplateGroup{
			TemplateID:      template.ID,
			Name:            groupReq.Name,
			Length:          groupReq.Length,
			Position:        i + 1,
			IsRequired:      groupReq.IsRequired,
			ValidationType:  groupReq.ValidationType,
			ValidationValue: groupReq.ValidationValue,
		}

		if err := s.templateRepo.CreateTemplateGroup(ctx, group); err != nil {
			return nil, fmt.Errorf("failed to create template group: %w", err)
		}
	}

	// Fetch the complete template with groups
	return s.templateRepo.GetByID(ctx, template.ID)
}

// DeleteTemplate deletes a template by ID with better error handling
func (s *GeneratorService) DeleteTemplate(ctx context.Context, id int64) error {
	// Check if template exists
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Delete the template - this will fail with a constraint error if there are dependencies
	if err := s.templateRepo.Delete(ctx, id); err != nil {
		// Parse and provide a better error message for foreign key constraint violation
		if strings.Contains(err.Error(), "foreign key constraint") ||
			strings.Contains(err.Error(), "violates foreign key constraint") {
			return fmt.Errorf("cannot delete template with associated hostnames: all hostnames using this template must be deleted first before the template can be removed (error: %w)", err)
		}
		return fmt.Errorf("failed to delete template: %w", err)
	}

	log.Info().
		Int64("id", id).
		Str("name", template.Name).
		Msg("Template deleted successfully")

	return nil
}
