package api

import (
	"net/http"
	"strconv"

	"github.com/bilbothegreedy/HNS/internal/dns"
	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// APIHandler handles API requests
type APIHandler struct {
	generatorService   *service.GeneratorService
	reservationService *service.ReservationService
	sequenceService    *service.SequenceService
	dnsChecker         *dns.DNSChecker
	dnsScanner         *dns.DNSScanner
}

// NewAPIHandler creates a new APIHandler
func NewAPIHandler(
	generatorService *service.GeneratorService,
	reservationService *service.ReservationService,
	sequenceService *service.SequenceService,
	dnsChecker *dns.DNSChecker,
) *APIHandler {
	dnsScanner := dns.NewDNSScanner(dnsChecker, generatorService)

	return &APIHandler{
		generatorService:   generatorService,
		reservationService: reservationService,
		sequenceService:    sequenceService,
		dnsChecker:         dnsChecker,
		dnsScanner:         dnsScanner,
	}
}

// HealthCheck handles health check requests
func (h *APIHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// GetTemplates handles requests to get all templates
func (h *APIHandler) GetTemplates(c *gin.Context) {
	// Parse pagination parameters
	limit, offset := getPaginationParams(c)

	// Get templates
	templates, total, err := h.generatorService.GetAvailableTemplates(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates"})
		log.Error().Err(err).Msg("Failed to get templates")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetTemplate handles requests to get a template by ID
func (h *APIHandler) GetTemplate(c *gin.Context) {
	// Parse template ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	// Get template
	template, err := h.generatorService.GetTemplateByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		log.Error().Err(err).Int64("templateID", id).Msg("Failed to get template")
		return
	}

	c.JSON(http.StatusOK, template)
}

// CreateTemplate handles requests to create a new template
func (h *APIHandler) CreateTemplate(c *gin.Context) {
	// Parse request
	var req models.TemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the authenticated user
	username, exists := c.Get("username")
	if !exists {
		apiKeyUserID, exists := c.Get("apiKeyUserID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User information not available"})
			return
		}
		username = "api-" + strconv.FormatInt(apiKeyUserID.(int64), 10)
	}
	req.CreatedBy = username.(string)

	// Create template
	template, err := h.generatorService.CreateTemplate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Error().Err(err).Msg("Failed to create template")
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GenerateHostname handles requests to generate a hostname
func (h *APIHandler) GenerateHostname(c *gin.Context) {
	// Parse request
	var req struct {
		TemplateID  int64             `json:"template_id" binding:"required"`
		SequenceNum int               `json:"sequence_num"`
		Params      map[string]string `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate hostname
	hostname, err := h.generatorService.GenerateHostname(c.Request.Context(), req.TemplateID, req.SequenceNum, req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Error().Err(err).Int64("templateID", req.TemplateID).Msg("Failed to generate hostname")
		return
	}

	// Check DNS if requested
	var dnsResult *models.DNSVerificationResult
	checkDNS := c.Query("check_dns") == "true"
	if checkDNS {
		dnsResult, err = h.dnsChecker.CheckHostname(c.Request.Context(), hostname)
		if err != nil {
			log.Warn().Err(err).Str("hostname", hostname).Msg("Failed to check hostname in DNS")
			// Continue without DNS check result
		}
	}

	// Build response
	response := gin.H{
		"hostname":     hostname,
		"template_id":  req.TemplateID,
		"sequence_num": req.SequenceNum,
		"params":       req.Params,
	}
	if dnsResult != nil {
		response["dns_check"] = dnsResult
	}

	c.JSON(http.StatusOK, response)
}

// ReserveHostname handles requests to reserve a hostname
func (h *APIHandler) ReserveHostname(c *gin.Context) {
	// Parse request
	var req models.HostnameReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the authenticated user
	username, exists := c.Get("username")
	if !exists {
		apiKeyUserID, exists := c.Get("apiKeyUserID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User information not available"})
			return
		}
		username = "api-" + strconv.FormatInt(apiKeyUserID.(int64), 10)
	}
	req.RequestedBy = username.(string)

	// Reserve hostname
	hostname, err := h.reservationService.ReserveHostname(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Error().Err(err).Int64("templateID", req.TemplateID).Msg("Failed to reserve hostname")
		return
	}

	c.JSON(http.StatusCreated, hostname)
}

// CommitHostname handles requests to commit a reserved hostname
func (h *APIHandler) CommitHostname(c *gin.Context) {
	// Parse request
	var req models.HostnameCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the authenticated user
	username, exists := c.Get("username")
	if !exists {
		apiKeyUserID, exists := c.Get("apiKeyUserID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User information not available"})
			return
		}
		username = "api-" + strconv.FormatInt(apiKeyUserID.(int64), 10)
	}
	req.CommittedBy = username.(string)

	// Commit hostname
	if err := h.reservationService.CommitHostname(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Error().Err(err).Int64("hostnameID", req.HostnameID).Msg("Failed to commit hostname")
		return
	}

	// Get updated hostname
	hostname, err := h.reservationService.GetHostname(c.Request.Context(), req.HostnameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated hostname"})
		log.Error().Err(err).Int64("hostnameID", req.HostnameID).Msg("Failed to get hostname after commit")
		return
	}

	c.JSON(http.StatusOK, hostname)
}

// ReleaseHostname handles requests to release a committed hostname
func (h *APIHandler) ReleaseHostname(c *gin.Context) {
	// Parse request
	var req models.HostnameReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the authenticated user
	username, exists := c.Get("username")
	if !exists {
		apiKeyUserID, exists := c.Get("apiKeyUserID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User information not available"})
			return
		}
		username = "api-" + strconv.FormatInt(apiKeyUserID.(int64), 10)
	}
	req.ReleasedBy = username.(string)

	// Release hostname
	if err := h.reservationService.ReleaseHostname(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Error().Err(err).Int64("hostnameID", req.HostnameID).Msg("Failed to release hostname")
		return
	}

	// Get updated hostname
	hostname, err := h.reservationService.GetHostname(c.Request.Context(), req.HostnameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated hostname"})
		log.Error().Err(err).Int64("hostnameID", req.HostnameID).Msg("Failed to get hostname after release")
		return
	}

	c.JSON(http.StatusOK, hostname)
}

// GetReservedHostnames handles requests to get all reserved hostnames
func (h *APIHandler) GetReservedHostnames(c *gin.Context) {
	// Parse pagination parameters
	limit, offset := getPaginationParams(c)

	// Get hostnames
	hostnames, err := h.reservationService.GetReservedHostnames(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reserved hostnames"})
		log.Error().Err(err).Msg("Failed to get reserved hostnames")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hostnames": hostnames,
		"count":     len(hostnames),
	})
}

// GetCommittedHostnames handles requests to get all committed hostnames
func (h *APIHandler) GetCommittedHostnames(c *gin.Context) {
	// Parse pagination parameters
	limit, offset := getPaginationParams(c)

	// Get hostnames
	hostnames, err := h.reservationService.GetCommittedHostnames(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get committed hostnames"})
		log.Error().Err(err).Msg("Failed to get committed hostnames")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hostnames": hostnames,
		"count":     len(hostnames),
	})
}

// GetHostname handles requests to get a hostname by ID
func (h *APIHandler) GetHostname(c *gin.Context) {
	// Parse hostname ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hostname ID"})
		return
	}

	// Get hostname
	hostname, err := h.reservationService.GetHostname(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hostname not found"})
		log.Error().Err(err).Int64("hostnameID", id).Msg("Failed to get hostname")
		return
	}

	c.JSON(http.StatusOK, hostname)
}

// CheckHostnameDNS handles requests to check a hostname in DNS
func (h *APIHandler) CheckHostnameDNS(c *gin.Context) {
	// Get hostname
	hostname := c.Param("hostname")
	if hostname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hostname is required"})
		return
	}

	// Check DNS
	result, err := h.dnsChecker.CheckHostname(c.Request.Context(), hostname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check hostname in DNS"})
		log.Error().Err(err).Str("hostname", hostname).Msg("Failed to check hostname in DNS")
		return
	}

	c.JSON(http.StatusOK, result)
}

// ScanDNS handles requests to scan DNS for hostnames
func (h *APIHandler) ScanDNS(c *gin.Context) {
	// Parse request
	var req dns.ScanOptions
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate options
	if req.TemplateID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}
	if req.StartSeq <= 0 {
		req.StartSeq = 1
	}
	if req.EndSeq <= 0 || req.EndSeq < req.StartSeq {
		req.EndSeq = req.StartSeq + 10 // Default to 10 hostnames
	}
	if req.MaxConcurrent <= 0 {
		req.MaxConcurrent = 10 // Default to 10 concurrent checks
	}

	// Scan DNS
	result, err := h.dnsScanner.ScanTemplate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan DNS"})
		log.Error().Err(err).Int64("templateID", req.TemplateID).Msg("Failed to scan DNS")
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetNextSequenceNumber handles requests to get the next sequence number for a template
func (h *APIHandler) GetNextSequenceNumber(c *gin.Context) {
	// Parse template ID
	templateIDStr := c.Param("templateID")
	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	// Get next sequence number
	nextSeq, err := h.sequenceService.GetNextSequenceNumber(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get next sequence number"})
		log.Error().Err(err).Int64("templateID", templateID).Msg("Failed to get next sequence number")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"template_id":  templateID,
		"sequence_num": nextSeq,
	})
}

// SearchHostnames handles requests to search for hostnames
func (h *APIHandler) SearchHostnames(c *gin.Context) {
	// Parse pagination parameters
	limit, offset := getPaginationParams(c)

	// Parse filters
	filters := make(map[string]interface{})

	// Template ID filter
	templateIDStr := c.Query("template_id")
	if templateIDStr != "" {
		templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
		if err == nil {
			filters["template_id"] = templateID
		}
	}

	// Status filter
	status := c.Query("status")
	if status != "" {
		filters["status"] = status
	}

	// Reserved by filter
	reservedBy := c.Query("reserved_by")
	if reservedBy != "" {
		filters["reserved_by"] = reservedBy
	}

	// Name filter (partial match)
	name := c.Query("name")
	if name != "" {
		filters["name LIKE"] = "%" + name + "%"
	}

	// Search hostnames
	hostnames, total, err := h.reservationService.SearchHostnames(c.Request.Context(), filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search hostnames"})
		log.Error().Err(err).Interface("filters", filters).Msg("Failed to search hostnames")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hostnames": hostnames,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// getPaginationParams extracts pagination parameters from the request
func getPaginationParams(c *gin.Context) (int, int) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}
