package handlers

import (
	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct {
	BaseHandler
	hostnameRepo repository.HostnameRepository
	templateRepo repository.TemplateRepository
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(
	hostnameRepo repository.HostnameRepository,
	templateRepo repository.TemplateRepository,
) *DashboardHandler {
	return &DashboardHandler{
		BaseHandler:  BaseHandler{},
		hostnameRepo: hostnameRepo,
		templateRepo: templateRepo,
	}
}

// Dashboard displays the dashboard page
func (h *DashboardHandler) Dashboard(c *gin.Context) {
	// If not logged in, redirect to login page
	loggedIn := helpers.IsAuthenticated(c)
	if !loggedIn {
		c.Redirect(302, "/login")
		return
	}

	// Get dashboard data
	ctx := c.Request.Context()

	// Count hostnames by status
	availableCount, _ := h.hostnameRepo.Count(ctx, 0, models.StatusAvailable)
	reservedCount, _ := h.hostnameRepo.Count(ctx, 0, models.StatusReserved)
	committedCount, _ := h.hostnameRepo.Count(ctx, 0, models.StatusCommitted)
	releasedCount, _ := h.hostnameRepo.Count(ctx, 0, models.StatusReleased)
	totalHostnames := availableCount + reservedCount + committedCount + releasedCount

	// Get recent hostnames
	filters := map[string]interface{}{}
	recentHostnames, _, err := h.hostnameRepo.List(ctx, 5, 0, filters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recent hostnames")
	}

	// Get template count
	templates, totalTemplates, err := h.templateRepo.List(ctx, 5, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get templates")
	}

	// Get user-specific data
	username, _ := c.Get("username")
	var userHostnames []*models.Hostname
	var userTotal int

	if username != nil {
		userFilters := map[string]interface{}{
			"reserved_by": username.(string),
		}
		userHostnames, userTotal, _ = h.hostnameRepo.List(ctx, 5, 0, userFilters)
	}

	// Render template
	h.RenderTemplate(c, "dashboard", gin.H{
		"Title":              "Dashboard",
		"ActivePage":         "dashboard",
		"TotalHostnames":     totalHostnames,
		"AvailableHostnames": availableCount,
		"ReservedHostnames":  reservedCount,
		"CommittedHostnames": committedCount,
		"ReleasedHostnames":  releasedCount,
		"RecentHostnames":    recentHostnames,
		"UserHostnames":      userHostnames,
		"UserHostnameCount":  userTotal,
		"TotalTemplates":     totalTemplates,
		"Templates":          templates,
	})
}

// Home redirects to the dashboard
func (h *DashboardHandler) Home(c *gin.Context) {
	// If not logged in, redirect to login page
	loggedIn := helpers.IsAuthenticated(c)
	if !loggedIn {
		c.Redirect(302, "/login")
		return
	}

	// Redirect to dashboard
	c.Redirect(302, "/dashboard")
}
