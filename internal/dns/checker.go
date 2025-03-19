package dns

import (
	"context"
	"fmt"
	"time"

	"github.com/bilbothegreedy/HNS/internal/config"
	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// DNSChecker is responsible for DNS-related operations
type DNSChecker struct {
	dnsConfig config.DNSConfig
	dnsClient *dns.Client
}

// NewDNSChecker creates a new DNSChecker
func NewDNSChecker(dnsConfig config.DNSConfig) *DNSChecker {
	return &DNSChecker{
		dnsConfig: dnsConfig,
		dnsClient: &dns.Client{
			Timeout: dnsConfig.Timeout,
		},
	}
}

// CheckHostname checks if a hostname exists in DNS
func (c *DNSChecker) CheckHostname(ctx context.Context, hostname string) (*models.DNSVerificationResult, error) {
	// Add domain suffix if needed
	if hostname == "" {
		return nil, fmt.Errorf("empty hostname")
	}

	// Create result object
	result := &models.DNSVerificationResult{
		Hostname:   hostname,
		Exists:     false,
		VerifiedAt: time.Now(),
	}

	// Create A record query
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(hostname), dns.TypeA)
	m.RecursionDesired = true

	// Try DNS servers in sequence
	var lastErr error
	for _, server := range c.dnsConfig.Servers {
		r, _, err := c.dnsClient.Exchange(m, server+":53")
		if err != nil {
			lastErr = err
			log.Warn().Err(err).Str("server", server).Str("hostname", hostname).Msg("DNS query failed")
			continue
		}

		// Check for successful response
		if r.Rcode == dns.RcodeSuccess {
			// Check if we got any answers
			if len(r.Answer) > 0 {
				result.Exists = true
				// Extract IP address if available
				for _, ans := range r.Answer {
					if a, ok := ans.(*dns.A); ok {
						result.IPAddress = a.A.String()
						break
					}
				}
			}
			return result, nil
		} else if r.Rcode == dns.RcodeNameError {
			// NXDOMAIN - domain definitely doesn't exist
			result.Exists = false
			return result, nil
		} else {
			// Other error, try next server
			lastErr = fmt.Errorf("DNS query returned error code: %d", r.Rcode)
			log.Warn().
				Int("rcode", r.Rcode).
				Str("server", server).
				Str("hostname", hostname).
				Msg("DNS query returned error code")
		}
	}

	// If we got here, all servers failed
	if lastErr != nil {
		return nil, fmt.Errorf("all DNS servers failed: %w", lastErr)
	}

	// Default to not exists if we couldn't determine
	return result, nil
}

// CheckMultipleHostnames checks multiple hostnames in parallel
func (c *DNSChecker) CheckMultipleHostnames(ctx context.Context, hostnames []string) ([]*models.DNSVerificationResult, error) {
	if len(hostnames) == 0 {
		return []*models.DNSVerificationResult{}, nil
	}

	// Create a results channel
	results := make(chan *models.DNSVerificationResult, len(hostnames))
	errors := make(chan error, len(hostnames))

	// Check each hostname in a goroutine
	for _, hostname := range hostnames {
		go func(h string) {
			result, err := c.CheckHostname(ctx, h)
			if err != nil {
				errors <- err
				results <- nil
			} else {
				results <- result
				errors <- nil
			}
		}(hostname)
	}

	// Collect results
	var verificationResults []*models.DNSVerificationResult
	var firstErr error
	for i := 0; i < len(hostnames); i++ {
		result := <-results
		err := <-errors
		if err != nil && firstErr == nil {
			firstErr = err
		}
		if result != nil {
			verificationResults = append(verificationResults, result)
		}
	}

	// Return results
	if len(verificationResults) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return verificationResults, nil
}
