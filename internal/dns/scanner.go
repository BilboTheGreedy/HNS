package dns

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/service"
	"github.com/rs/zerolog/log"
)

// DNSScanner is responsible for scanning DNS for hostnames
type DNSScanner struct {
	dnsChecker   *DNSChecker
	generatorSvc *service.GeneratorService
}

// NewDNSScanner creates a new DNSScanner
func NewDNSScanner(dnsChecker *DNSChecker, generatorSvc *service.GeneratorService) *DNSScanner {
	return &DNSScanner{
		dnsChecker:   dnsChecker,
		generatorSvc: generatorSvc,
	}
}

// ScanResult represents the result of a DNS scan
type ScanResult struct {
	TemplateID        int64      `json:"template_id"`
	TemplateName      string     `json:"template_name"`
	TotalHostnames    int        `json:"total_hostnames"`
	ExistingHostnames int        `json:"existing_hostnames"`
	ScanDuration      string     `json:"scan_duration"`
	Results           []ScanItem `json:"results"`
}

// ScanItem represents a single hostname scan result
type ScanItem struct {
	Hostname  string `json:"hostname"`
	Exists    bool   `json:"exists"`
	IPAddress string `json:"ip_address,omitempty"`
}

// ScanOptions represents options for a DNS scan
type ScanOptions struct {
	TemplateID    int64             `json:"template_id"`
	StartSeq      int               `json:"start_seq"`
	EndSeq        int               `json:"end_seq"`
	Params        map[string]string `json:"params"`
	MaxConcurrent int               `json:"max_concurrent"`
}

// ScanTemplate scans DNS for hostnames based on a template
func (s *DNSScanner) ScanTemplate(ctx context.Context, options ScanOptions) (*ScanResult, error) {
	startTime := time.Now()

	// Validate options
	if options.TemplateID <= 0 {
		return nil, fmt.Errorf("invalid template ID")
	}
	if options.EndSeq < options.StartSeq {
		return nil, fmt.Errorf("end sequence must be greater than or equal to start sequence")
	}
	if options.MaxConcurrent <= 0 {
		options.MaxConcurrent = 10 // Default to 10 concurrent checks
	}

	// Get template
	template, err := s.generatorSvc.GetTemplateByID(ctx, options.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Initialize result
	result := &ScanResult{
		TemplateID:   options.TemplateID,
		TemplateName: template.Name,
		Results:      []ScanItem{},
	}

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, options.MaxConcurrent)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex

	// Generate and check hostnames for each sequence number
	for seq := options.StartSeq; seq <= options.EndSeq; seq++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(sequenceNum int) {
			defer func() {
				<-sem // Release semaphore
				wg.Done()
			}()

			// Generate hostname
			hostname, err := s.generatorSvc.GenerateHostname(ctx, options.TemplateID, sequenceNum, options.Params)
			if err != nil {
				log.Error().Err(err).Int("sequence", sequenceNum).Msg("Failed to generate hostname")
				return
			}

			// Check if hostname exists in DNS
			dnsResult, err := s.dnsChecker.CheckHostname(ctx, hostname)
			if err != nil {
				log.Error().Err(err).Str("hostname", hostname).Msg("Failed to check hostname in DNS")
				return
			}

			// Add to results
			resultsMutex.Lock()
			result.Results = append(result.Results, ScanItem{
				Hostname:  hostname,
				Exists:    dnsResult.Exists,
				IPAddress: dnsResult.IPAddress,
			})
			if dnsResult.Exists {
				result.ExistingHostnames++
			}
			resultsMutex.Unlock()
		}(seq)
	}

	// Wait for all checks to complete
	wg.Wait()
	result.TotalHostnames = len(result.Results)
	result.ScanDuration = time.Since(startTime).String()

	return result, nil
}

// DiscoverSequenceRange attempts to discover the range of sequence numbers in use
func (s *DNSScanner) DiscoverSequenceRange(ctx context.Context, templateID int64, params map[string]string) (int, int, error) {
	// Get template
	template, err := s.generatorSvc.GetTemplateByID(ctx, templateID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get template: %w", err)
	}

	// Start with the template's sequence start
	startSeq := template.SequenceStart

	// Check if the first few hostnames exist to find a starting point
	var lowestFound, highestFound int

	// Binary search to find the lower bound
	low := startSeq
	high := startSeq + 1000 // Arbitrary limit

	// First do a quick scan to find if there are any hostnames at all
	foundAny := false
	for i := low; i <= low+10; i++ {
		hostname, err := s.generatorSvc.GenerateHostname(ctx, templateID, i, params)
		if err != nil {
			continue
		}

		result, err := s.dnsChecker.CheckHostname(ctx, hostname)
		if err != nil {
			continue
		}

		if result.Exists {
			foundAny = true
			lowestFound = i
			highestFound = i
			break
		}
	}

	if !foundAny {
		// Try a wider range
		for i := low; i <= high; i += 100 {
			hostname, err := s.generatorSvc.GenerateHostname(ctx, templateID, i, params)
			if err != nil {
				continue
			}

			result, err := s.dnsChecker.CheckHostname(ctx, hostname)
			if err != nil {
				continue
			}

			if result.Exists {
				foundAny = true
				lowestFound = i
				highestFound = i
				break
			}
		}
	}

	if !foundAny {
		return 0, 0, fmt.Errorf("no existing hostnames found for template")
	}

	// Now find lower bound, starting from the found point and going down
	for i := lowestFound - 1; i >= startSeq; i-- {
		hostname, err := s.generatorSvc.GenerateHostname(ctx, templateID, i, params)
		if err != nil {
			break
		}

		result, err := s.dnsChecker.CheckHostname(ctx, hostname)
		if err != nil || !result.Exists {
			break
		}

		lowestFound = i
	}

	// Find upper bound, starting from found point and going up
	for i := highestFound + 1; i <= highestFound+1000; i++ {
		hostname, err := s.generatorSvc.GenerateHostname(ctx, templateID, i, params)
		if err != nil {
			break
		}

		result, err := s.dnsChecker.CheckHostname(ctx, hostname)
		if err != nil || !result.Exists {
			// If we get 10 consecutive non-existent hostnames, assume we've reached the end
			consecutive := 1
			for j := i + 1; j <= i+10; j++ {
				hostname, err := s.generatorSvc.GenerateHostname(ctx, templateID, j, params)
				if err != nil {
					continue
				}

				result, err := s.dnsChecker.CheckHostname(ctx, hostname)
				if err != nil || !result.Exists {
					consecutive++
				} else {
					break
				}
			}

			if consecutive >= 10 {
				break
			}
		}

		highestFound = i
	}

	return lowestFound, highestFound, nil
}

// ParseHostnameSequence attempts to extract the sequence number from a hostname
func (s *DNSScanner) ParseHostnameSequence(hostname string, template *models.Template) (int, error) {
	// This is a simplified implementation that assumes the sequence is numeric
	// and is at the end of the hostname

	// Extract the trailing digits
	digits := ""
	for i := len(hostname) - 1; i >= 0; i-- {
		if hostname[i] >= '0' && hostname[i] <= '9' {
			digits = string(hostname[i]) + digits
		} else {
			break
		}
	}

	if digits == "" {
		return 0, fmt.Errorf("no sequence number found in hostname")
	}

	// Convert to integer
	seq, err := strconv.Atoi(digits)
	if err != nil {
		return 0, fmt.Errorf("failed to parse sequence number: %w", err)
	}

	return seq, nil
}

// AnalyzeTemplateUsage analyzes DNS to determine hostname usage patterns
func (s *DNSScanner) AnalyzeTemplateUsage(ctx context.Context, templateID int64, sampleSize int) (map[string]int, error) {
	// Get template
	template, err := s.generatorSvc.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Check if template has any groups
	if len(template.Groups) == 0 {
		return nil, fmt.Errorf("template has no groups")
	}

	// Initialize result
	usage := make(map[string]int)

	// Find a range of sequence numbers to check
	low, high, err := s.DiscoverSequenceRange(ctx, templateID, nil)
	if err != nil {
		// If we can't find a range, use default range
		low = template.SequenceStart
		high = low + sampleSize - 1
	}

	// If range is too large, limit it
	if high-low+1 > sampleSize {
		high = low + sampleSize - 1
	}

	// Check hostnames in the range
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, 10)

	for seq := low; seq <= high; seq++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(sequenceNum int) {
			defer func() {
				<-sem // Release semaphore
				wg.Done()
			}()

			// Generate hostname
			hostname, err := s.generatorSvc.GenerateHostname(ctx, templateID, sequenceNum, nil)
			if err != nil {
				return
			}

			// Check if hostname exists in DNS
			dnsResult, err := s.dnsChecker.CheckHostname(ctx, hostname)
			if err != nil || !dnsResult.Exists {
				return
			}

			// Analyze the hostname components
			// This is a simplified analysis that looks at the first few characters
			prefix := ""
			if len(hostname) >= 2 {
				prefix = strings.ToUpper(hostname[:2])
			}

			mutex.Lock()
			usage[prefix]++
			mutex.Unlock()
		}(seq)
	}

	// Wait for all checks to complete
	wg.Wait()

	return usage, nil
}
