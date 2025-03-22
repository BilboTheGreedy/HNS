package handlers

import (
	"fmt"
	"strconv"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/service"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// HostnameHandler handles hostname-related requests
type HostnameHandler struct {
	BaseHandler
	hostnameRepo   repository.HostnameRepository
	templateRepo   repository.TemplateRepository
	generatorSvc   *service.GeneratorService
	reservationSvc *service.ReservationService
}

// NewHostnameHandler creates a new HostnameHandler
func NewHostnameHandler(
	hostnameRepo repository.HostnameRepository,
	templateRepo repository.TemplateRepository,
	generatorSvc *service.GeneratorService,
	reservationSvc *service.ReservationService,
) *HostnameHandler {
	return &HostnameHandler{
		BaseHandler:    BaseHandler{},
		hostnameRepo:   hostnameRepo,
		templateRepo:   templateRepo,
		generatorSvc:   generatorSvc,
		reservationSvc: reservationSvc,
	}
}

// HostnameList displays the list of hostnames
func (h *HostnameHandler) HostnameList(c *gin.Context) {
	// Get pagination parameters
	limit, offset := helpers.GetPaginationParams(c)

	// Get filter parameters
	filters := map[string]interface{}{}

	// Apply name filter if provided
	name := c.Query("name")
	if name != "" {
		filters["name LIKE"] = "%" + name + "%"
	}

	// Apply template filter if provided
	templateIDStr := c.Query("template_id")
	if templateIDStr != "" {
		templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
		if err == nil {
			filters["template_id"] = templateID
		}
	}

	// Apply status filter if provided
	status := c.Query("status")
	if status != "" {
		filters["status"] = status
	}

	// Apply reserved_by filter if provided
	reservedBy := c.Query("reserved_by")
	if reservedBy != "" {
		filters["reserved_by"] = reservedBy
	}

	// Get hostnames with filters
	ctx := c.Request.Context()
	hostnames, total, err := h.hostnameRepo.List(ctx, limit, offset, filters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get hostnames")
		h.RedirectWithAlert(c, "/hostnames", "danger", "Error retrieving hostnames")
		return
	}

	// Get templates for filter dropdown and display
	templates, _, _ := h.templateRepo.List(ctx, 100, 0)

	// Get template names for each hostname
	hostnamesWithTemplates := make([]gin.H, len(hostnames))
	for i, hostname := range hostnames {
		templateName := "Unknown"
		for _, tmpl := range templates {
			if tmpl.ID == hostname.TemplateID {
				templateName = tmpl.Name
				break
			}
		}

		hostnamesWithTemplates[i] = gin.H{
			"ID":           hostname.ID,
			"Name":         hostname.Name,
			"TemplateID":   hostname.TemplateID,
			"TemplateName": templateName,
			"Status":       hostname.Status,
			"SequenceNum":  hostname.SequenceNum,
			"ReservedBy":   hostname.ReservedBy,
			"ReservedAt":   hostname.ReservedAt,
		}
	}

	// Build pagination URL
	paginationURL := "/hostnames?"
	if name != "" {
		paginationURL += "name=" + name + "&"
	}
	if templateIDStr != "" {
		paginationURL += "template_id=" + templateIDStr + "&"
	}
	if status != "" {
		paginationURL += "status=" + status + "&"
	}
	if reservedBy != "" {
		paginationURL += "reserved_by=" + reservedBy + "&"
	}

	// Combine pagination data with template data
	templateData := gin.H{
		"Title":      "Hostnames",
		"ActivePage": "hostnames",
		"Hostnames":  hostnamesWithTemplates,
		"Templates":  templates,
		"Filters": gin.H{
			"name":        name,
			"template_id": templateIDStr,
			"status":      status,
			"reserved_by": reservedBy,
		},
		"PaginationURL": paginationURL,
		"Total":         total,
		"Limit":         limit,
		"Offset":        offset,
	}

	// Add pagination data
	paginationData := helpers.GetPaginationData(total, limit, offset)
	for k, v := range paginationData {
		templateData[k] = v
	}

	h.RenderTemplate(c, "hostname_list", templateData)
}

// HostnameDetail displays the details of a hostname
func (h *HostnameHandler) HostnameDetail(c *gin.Context) {
	// Get hostname ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/hostnames", "danger", "Invalid hostname ID")
		return
	}

	// Get hostname
	ctx := c.Request.Context()
	hostname, err := h.hostnameRepo.GetByID(ctx, id)
	if err != nil {
		h.RedirectWithAlert(c, "/hostnames", "danger", "Hostname not found")
		return
	}

	// Get template
	template, err := h.templateRepo.GetByID(ctx, hostname.TemplateID)
	if err != nil {
		template = &models.Template{
			Name: "Unknown",
		}
	}

	// Render template
	h.RenderTemplate(c, "hostname_detail", gin.H{
		"Title":      hostname.Name,
		"ActivePage": "hostnames",
		"Hostname":   hostname,
		"Template":   template,
	})
}

// GenerateHostnamePage displays the hostname generation form
func (h *HostnameHandler) GenerateHostnamePage(c *gin.Context) {
	// Get templates
	ctx := c.Request.Context()
	templates, _, err := h.templateRepo.List(ctx, 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get templates")
		h.RedirectWithAlert(c, "/hostnames", "danger", "Error retrieving templates")
		return
	}

	// Render template
	h.RenderTemplate(c, "hostname_generate", gin.H{
		"Title":      "Generate Hostname",
		"ActivePage": "hostnames",
		"Templates":  templates,
	})
}

// GenerateHostname handles hostname generation form submission
func (h *HostnameHandler) GenerateHostname(c *gin.Context) {
	// Get form data
	templateIDStr := c.PostForm("template_id")
	sequenceNumStr := c.PostForm("sequence_num")
	//checkDNS := c.PostForm("check_dns") == "on"

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/hostnames/generate", "danger", "Invalid template ID")
		return
	}

	// Parse sequence number if provided
	var sequenceNum int
	if sequenceNumStr != "" {
		sequenceNum, err = strconv.Atoi(sequenceNumStr)
		if err != nil {
			sequenceNum = 0 // Use default
		}
	}

	// Get template parameters dynamically
	params := make(map[string]string)
	// Get template to find expected parameters
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

	// Generate hostname
	hostname, err := h.generatorSvc.GenerateHostname(c.Request.Context(), templateID, sequenceNum, params)
	if err != nil {
		h.RedirectWithAlert(c, "/hostnames/generate", "danger", "Failed to generate hostname: "+err.Error())
		return
	}

	// If check DNS is enabled, implement DNS check here
	// For now, just redirect to hostname page
	h.RedirectWithAlert(c, "/hostnames/generate", "success", "Hostname generated: "+hostname)
}

// ReserveHostnamePage displays the hostname reservation form
func (h *HostnameHandler) ReserveHostnamePage(c *gin.Context) {
	// Get templates
	ctx := c.Request.Context()
	templates, _, err := h.templateRepo.List(ctx, 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get templates")
		h.RedirectWithAlert(c, "/hostnames", "danger", "Error retrieving templates")
		return
	}

	// Render template
	h.RenderTemplate(c, "hostname_reserve", gin.H{
		"Title":      "Reserve Hostname",
		"ActivePage": "hostnames",
		"Templates":  templates,
	})
}

// ReserveHostname handles hostname reservation form submission
func (h *HostnameHandler) ReserveHostname(c *gin.Context) {
	// Get form data
	templateIDStr := c.PostForm("template_id")

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/hostnames/reserve", "danger", "Invalid template ID")
		return
	}

	// Get username
	username, exists := c.Get("username")
	if !exists {
		username = "unknown"
	}

	// Get template parameters dynamically
	params := make(map[string]string)
	// Get template to find expected parameters
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

	// Create reservation request
	req := &models.HostnameReservationRequest{
		TemplateID:  templateID,
		Params:      params,
		RequestedBy: username.(string),
	}

	// Reserve hostname
	hostname, err := h.reservationSvc.ReserveHostname(c.Request.Context(), req)
	if err != nil {
		h.RedirectWithAlert(c, "/hostnames/reserve", "danger", "Failed to reserve hostname: "+err.Error())
		return
	}

	// Redirect to hostname detail
	h.RedirectWithAlert(c, fmt.Sprintf("/hostnames/%d", hostname.ID), "success", "Hostname reserved successfully")
}

// CommitHostname commits a reserved hostname
func (h *HostnameHandler) CommitHostname(c *gin.Context) {
	// Get hostname ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/hostnames", "danger", "Invalid hostname ID")
		return
	}

	// Get username
	username, exists := c.Get("username")
	if !exists {
		username = "unknown"
	}

	// Create commit request
	req := &models.HostnameCommitRequest{
		HostnameID:  id,
		CommittedBy: username.(string),
	}

	// Commit hostname
	err = h.reservationSvc.CommitHostname(c.Request.Context(), req)
	if err != nil {
		h.RedirectWithAlert(c, fmt.Sprintf("/hostnames/%d", id), "danger", "Failed to commit hostname: "+err.Error())
		return
	}

	// Redirect to hostname detail
	h.RedirectWithAlert(c, fmt.Sprintf("/hostnames/%d", id), "success", "Hostname committed successfully")
}

// ReleaseHostname releases a committed hostname
func (h *HostnameHandler) ReleaseHostname(c *gin.Context) {
	// Get hostname ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/hostnames", "danger", "Invalid hostname ID")
		return
	}

	// Get username
	username, exists := c.Get("username")
	if !exists {
		username = "unknown"
	}

	// Create release request
	req := &models.HostnameReleaseRequest{
		HostnameID: id,
		ReleasedBy: username.(string),
	}

	// Release hostname
	err = h.reservationSvc.ReleaseHostname(c.Request.Context(), req)
	if err != nil {
		h.RedirectWithAlert(c, fmt.Sprintf("/hostnames/%d", id), "danger", "Failed to release hostname: "+err.Error())
		return
	}

	// Redirect to hostname detail
	h.RedirectWithAlert(c, fmt.Sprintf("/hostnames/%d", id), "success", "Hostname released successfully")
}
