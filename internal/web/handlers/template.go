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

// TemplateHandler handles template-related requests
type TemplateHandler struct {
	BaseHandler
	templateRepo repository.TemplateRepository
	generatorSvc *service.GeneratorService
}

// NewTemplateHandler creates a new TemplateHandler
func NewTemplateHandler(
	templateRepo repository.TemplateRepository,
	generatorSvc *service.GeneratorService,
) *TemplateHandler {
	return &TemplateHandler{
		BaseHandler:  BaseHandler{},
		templateRepo: templateRepo,
		generatorSvc: generatorSvc,
	}
}

// TemplateList displays the list of templates
func (h *TemplateHandler) TemplateList(c *gin.Context) {
	// Get pagination parameters
	limit, offset := helpers.GetPaginationParams(c)

	// Get templates
	ctx := c.Request.Context()
	templates, total, err := h.templateRepo.List(ctx, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get templates")
		h.RenderTemplate(c, "template_list", gin.H{
			"Title":     "Templates",
			"Templates": []*models.Template{},
		})
		return
	}

	// Combine pagination data with template data
	templateData := gin.H{
		"Title":      "Templates",
		"ActivePage": "templates",
		"Templates":  templates,
	}

	// Add pagination data
	paginationData := helpers.GetPaginationData(total, limit, offset)
	for k, v := range paginationData {
		templateData[k] = v
	}

	h.RenderTemplate(c, "template_list", templateData)
}

// TemplateDetail displays the details of a template
func (h *TemplateHandler) TemplateDetail(c *gin.Context) {
	// Get template ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/templates", "danger", "Invalid template ID")
		return
	}

	// Get template
	ctx := c.Request.Context()
	template, err := h.templateRepo.GetByID(ctx, id)
	if err != nil {
		h.RedirectWithAlert(c, "/templates", "danger", "Template not found")
		return
	}

	// Render template
	h.RenderTemplate(c, "template_detail", gin.H{
		"Title":      template.Name,
		"ActivePage": "templates",
		"Template":   template,
	})
}

// NewTemplate displays the template creation form
func (h *TemplateHandler) NewTemplate(c *gin.Context) {
	// Check if user is admin
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Render template
	h.RenderTemplate(c, "template_form", gin.H{
		"Title":      "Create Template",
		"ActivePage": "templates",
		"Template":   &models.Template{},
		"IsNew":      true,
	})
}

// CreateTemplate handles template creation form submission
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	// Check if user is admin
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Get username
	username, _ := c.Get("username")
	if username == nil {
		username = "unknown"
	}

	// Get form data
	name := c.PostForm("name")
	description := c.PostForm("description")
	maxLengthStr := c.PostForm("max_length")
	sequenceStartStr := c.PostForm("sequence_start")
	sequenceLengthStr := c.PostForm("sequence_length")
	sequencePadding := c.PostForm("sequence_padding") == "on"
	sequenceIncrementStr := c.PostForm("sequence_increment")
	//isActive := c.PostForm("is_active") == "on"

	// Parse numeric values
	maxLength, _ := strconv.Atoi(maxLengthStr)
	sequenceStart, _ := strconv.Atoi(sequenceStartStr)
	sequenceLength, _ := strconv.Atoi(sequenceLengthStr)
	sequenceIncrement, _ := strconv.Atoi(sequenceIncrementStr)

	// Handle default values
	if maxLength <= 0 {
		maxLength = 15
	}
	if sequenceStart <= 0 {
		sequenceStart = 1
	}
	if sequenceLength <= 0 {
		sequenceLength = 3
	}
	if sequenceIncrement <= 0 {
		sequenceIncrement = 1
	}

	// Create template request
	req := &models.TemplateCreateRequest{
		Name:              name,
		Description:       description,
		MaxLength:         maxLength,
		SequenceStart:     sequenceStart,
		SequenceLength:    sequenceLength,
		SequencePadding:   sequencePadding,
		SequenceIncrement: sequenceIncrement,
		CreatedBy:         username.(string),
		Groups:            []models.TemplateGroupRequest{},
	}

	// Process groups
	groupCount, err := strconv.Atoi(c.PostForm("group_count"))
	if err != nil || groupCount <= 0 {
		// Try to process groups dynamically
		i := 0
		for {
			nameKey := fmt.Sprintf("groups[%d][name]", i)
			groupName := c.PostForm(nameKey)
			if groupName == "" {
				break // No more groups
			}

			// Get group data
			lengthStr := c.PostForm(fmt.Sprintf("groups[%d][length]", i))
			validationType := c.PostForm(fmt.Sprintf("groups[%d][validation_type]", i))
			validationValue := c.PostForm(fmt.Sprintf("groups[%d][validation_value]", i))
			isRequired := c.PostForm(fmt.Sprintf("groups[%d][is_required]", i)) == "on"

			// Parse length
			length, _ := strconv.Atoi(lengthStr)
			if length <= 0 {
				length = 1
			}

			// Add group
			group := models.TemplateGroupRequest{
				Name:            groupName,
				Length:          length,
				IsRequired:      isRequired,
				ValidationType:  validationType,
				ValidationValue: validationValue,
			}
			req.Groups = append(req.Groups, group)

			i++
		}
	}

	// Validate template
	if len(req.Groups) == 0 {
		h.RenderTemplate(c, "template_form", gin.H{
			"Title":      "Create Template",
			"ActivePage": "templates",
			"Template":   req,
			"IsNew":      true,
			"Error":      "Template must have at least one group",
		})
		return
	}

	// Create template
	template, err := h.generatorSvc.CreateTemplate(c.Request.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create template")
		h.RenderTemplate(c, "template_form", gin.H{
			"Title":      "Create Template",
			"ActivePage": "templates",
			"Template":   req,
			"IsNew":      true,
			"Error":      "Failed to create template: " + err.Error(),
		})
		return
	}

	// Success
	h.RedirectWithAlert(c, "/templates/"+strconv.FormatInt(template.ID, 10), "success", "Template created successfully")
}

// EditTemplate displays the template edit form
func (h *TemplateHandler) EditTemplate(c *gin.Context) {
	// Check if user is admin
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Get template ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/templates", "danger", "Invalid template ID")
		return
	}

	// Get template
	ctx := c.Request.Context()
	template, err := h.templateRepo.GetByID(ctx, id)
	if err != nil {
		h.RedirectWithAlert(c, "/templates", "danger", "Template not found")
		return
	}

	// Render template
	h.RenderTemplate(c, "template_form", gin.H{
		"Title":      "Edit Template",
		"ActivePage": "templates",
		"Template":   template,
		"IsNew":      false,
	})
}

// UpdateTemplate handles template update form submission
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	// Check if user is admin
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Get template ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/templates", "danger", "Invalid template ID")
		return
	}

	// Get username
	username, _ := c.Get("username")
	if username == nil {
		username = "unknown"
	}

	// Get template to update
	ctx := c.Request.Context()
	template, err := h.templateRepo.GetByID(ctx, id)
	if err != nil {
		h.RedirectWithAlert(c, "/templates", "danger", "Template not found")
		return
	}

	// Get form data
	name := c.PostForm("name")
	description := c.PostForm("description")
	maxLengthStr := c.PostForm("max_length")
	sequenceStartStr := c.PostForm("sequence_start")
	sequenceLengthStr := c.PostForm("sequence_length")
	sequencePadding := c.PostForm("sequence_padding") == "on"
	sequenceIncrementStr := c.PostForm("sequence_increment")
	isActive := c.PostForm("is_active") == "on"

	// Parse numeric values
	maxLength, _ := strconv.Atoi(maxLengthStr)
	sequenceStart, _ := strconv.Atoi(sequenceStartStr)
	sequenceLength, _ := strconv.Atoi(sequenceLengthStr)
	sequenceIncrement, _ := strconv.Atoi(sequenceIncrementStr)

	// Update template fields
	template.Name = name
	template.Description = description
	template.MaxLength = maxLength
	template.SequenceStart = sequenceStart
	template.SequenceLength = sequenceLength
	template.SequencePadding = sequencePadding
	template.SequenceIncrement = sequenceIncrement
	template.IsActive = isActive

	// Save changes
	err = h.templateRepo.Update(ctx, template)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update template")
		h.RenderTemplate(c, "template_form", gin.H{
			"Title":      "Edit Template",
			"ActivePage": "templates",
			"Template":   template,
			"IsNew":      false,
			"Error":      "Failed to update template: " + err.Error(),
		})
		return
	}

	// Note: Groups would need to be updated via separate API calls
	// For simplicity, we're only updating the base template here

	// Success
	h.RedirectWithAlert(c, "/templates/"+strconv.FormatInt(template.ID, 10), "success", "Template updated successfully")
}

// DeleteTemplate handles template deletion
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	// Check if user is admin
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Get template ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/templates", "danger", "Invalid template ID")
		return
	}

	// Delete template
	err = h.templateRepo.Delete(c.Request.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete template")
		h.RedirectWithAlert(c, "/templates", "danger", "Failed to delete template: "+err.Error())
		return
	}

	// Success
	h.RedirectWithAlert(c, "/templates", "success", "Template deleted successfully")
}
