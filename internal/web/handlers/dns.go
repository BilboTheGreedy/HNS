package handlers

import (
	"context"
	"strconv"

	"github.com/bilbothegreedy/HNS/internal/dns"
	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// DNSHandler handles DNS-related requests
type DNSHandler struct {
	BaseHandler
	templateRepo repository.TemplateRepository
	dnsChecker   interface{} // Using interface for flexibility
	dnsScanner   *dns.DNSScanner
}

// NewDNSHandler creates a new DNSHandler
func NewDNSHandler(
	templateRepo repository.TemplateRepository,
	dnsChecker interface{},
	generatorSvc *service.GeneratorService,
) *DNSHandler {
	var scanner *dns.DNSScanner

	// If dnsChecker implements the expected interface, create the scanner
	if checker, ok := dnsChecker.(*dns.DNSChecker); ok {
		scanner = dns.NewDNSScanner(checker, generatorSvc)
	}

	return &DNSHandler{
		BaseHandler:  BaseHandler{},
		templateRepo: templateRepo,
		dnsChecker:   dnsChecker,
		dnsScanner:   scanner,
	}
}

// DNSTools displays the DNS tools landing page
func (h *DNSHandler) DNSTools(c *gin.Context) {
	h.RenderTemplate(c, "dns_tools", gin.H{
		"Title":      "DNS Tools",
		"ActivePage": "dns",
	})
}

// DNSCheck displays the DNS check form
func (h *DNSHandler) DNSCheck(c *gin.Context) {
	h.RenderTemplate(c, "dns_check", gin.H{
		"Title":      "DNS Check",
		"ActivePage": "dns",
	})
}

// DNSCheckHostname checks a specific hostname and displays results
func (h *DNSHandler) DNSCheckHostname(c *gin.Context) {
	// Get hostname from path
	hostname := c.Param("hostname")
	if hostname == "" {
		h.RedirectWithAlert(c, "/dns/check", "danger", "Hostname is required")
		return
	}

	// Check DNS if checker is available
	var dnsResult *models.DNSVerificationResult
	if checker, ok := h.dnsChecker.(interface {
		CheckHostname(ctx context.Context, hostname string) (*models.DNSVerificationResult, error)
	}); ok {
		result, err := checker.CheckHostname(c.Request.Context(), hostname)
		if err != nil {
			log.Error().Err(err).Str("hostname", hostname).Msg("Failed to check hostname in DNS")
		} else {
			dnsResult = result
		}
	}

	// Render template
	h.RenderTemplate(c, "dns_check", gin.H{
		"Title":      "DNS Check",
		"ActivePage": "dns",
		"Hostname":   hostname,
		"DNSResult":  dnsResult,
	})
}

// DNSScan displays the DNS scan form
func (h *DNSHandler) DNSScan(c *gin.Context) {
	// Get templates
	ctx := c.Request.Context()
	templates, _, err := h.templateRepo.List(ctx, 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get templates")
	}

	// Render template
	h.RenderTemplate(c, "dns_scan", gin.H{
		"Title":      "DNS Scan",
		"ActivePage": "dns",
		"Templates":  templates,
	})
}

// DNSScanSubmit performs a DNS scan
func (h *DNSHandler) DNSScanSubmit(c *gin.Context) {
	// Check if scanner is available
	if h.dnsScanner == nil {
		h.RedirectWithAlert(c, "/dns/scan", "danger", "DNS scanner not available")
		return
	}

	// Get form data
	templateIDStr := c.PostForm("template_id")
	startSeqStr := c.PostForm("start_seq")
	endSeqStr := c.PostForm("end_seq")
	maxConcurrentStr := c.PostForm("max_concurrent")

	// Parse template ID
	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/dns/scan", "danger", "Invalid template ID")
		return
	}

	// Parse start and end sequence
	startSeq, err := strconv.Atoi(startSeqStr)
	if err != nil || startSeq <= 0 {
		startSeq = 1
	}

	endSeq, err := strconv.Atoi(endSeqStr)
	if err != nil || endSeq <= 0 {
		endSeq = startSeq + 10 // Default 10 hostnames
	}

	// Parse max concurrent
	maxConcurrent, err := strconv.Atoi(maxConcurrentStr)
	if err != nil || maxConcurrent <= 0 {
		maxConcurrent = 10 // Default 10 concurrent
	}

	// Get template parameters
	params := make(map[string]string)
	template, err := h.templateRepo.GetByID(c.Request.Context(), templateID)
	if err == nil {
		for _, group := range template.Groups {
			// Skip sequence groups
			if group.ValidationType == string(models.ValidationTypeSequence) {
				continue
			}

			// Get parameter value
			paramName := "param_" + group.Name
			paramValue := c.PostForm(paramName)
			if paramValue != "" {
				params[group.Name] = paramValue
			}
		}
	}

	// Create scan options
	options := dns.ScanOptions{
		TemplateID:    templateID,
		StartSeq:      startSeq,
		EndSeq:        endSeq,
		Params:        params,
		MaxConcurrent: maxConcurrent,
	}

	// Perform scan
	result, err := h.dnsScanner.ScanTemplate(c.Request.Context(), options)
	if err != nil {
		log.Error().Err(err).Msg("Failed to scan DNS")
		h.RedirectWithAlert(c, "/dns/scan", "danger", "Failed to scan DNS: "+err.Error())
		return
	}

	// Render template with results
	h.RenderTemplate(c, "dns_scan_results", gin.H{
		"Title":      "DNS Scan Results",
		"ActivePage": "dns",
		"ScanResult": result,
	})
}

// DNSDiscover attempts to discover the range of sequence numbers in use
func (h *DNSHandler) DNSDiscover(c *gin.Context) {
	// Check if scanner is available
	if h.dnsScanner == nil {
		h.RedirectWithAlert(c, "/dns/scan", "danger", "DNS scanner not available")
		return
	}

	// Get form data
	templateIDStr := c.PostForm("template_id")

	// Parse template ID
	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/dns/scan", "danger", "Invalid template ID")
		return
	}

	// Get template parameters
	params := make(map[string]string)
	template, err := h.templateRepo.GetByID(c.Request.Context(), templateID)
	if err == nil {
		for _, group := range template.Groups {
			// Skip sequence groups
			if group.ValidationType == string(models.ValidationTypeSequence) {
				continue
			}

			// Get parameter value
			paramName := "param_" + group.Name
			paramValue := c.PostForm(paramName)
			if paramValue != "" {
				params[group.Name] = paramValue
			}
		}
	}

	// Discover sequence range
	low, high, err := h.dnsScanner.DiscoverSequenceRange(c.Request.Context(), templateID, params)
	if err != nil {
		log.Error().Err(err).Msg("Failed to discover sequence range")
		h.RedirectWithAlert(c, "/dns/scan", "danger", "Failed to discover sequence range: "+err.Error())
		return
	}

	// Set discovered range in the form
	h.RenderTemplate(c, "dns_scan", gin.H{
		"Title":            "DNS Scan",
		"ActivePage":       "dns",
		"Templates":        []*models.Template{template},
		"TemplateID":       templateID,
		"StartSeq":         low,
		"EndSeq":           high,
		"Params":           params,
		"Discovered":       true,
		"DiscoveryMessage": "Discovered hostname sequence range: " + strconv.Itoa(low) + " to " + strconv.Itoa(high),
	})
}
