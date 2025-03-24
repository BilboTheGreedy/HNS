package handlers

import (
	"context"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/gin-gonic/gin"
)

// DashboardStats represents statistics for the dashboard
type DashboardStats struct {
	TotalHostnames     int `json:"total_hostnames"`
	ReservedHostnames  int `json:"reserved_hostnames"`
	AvailableHostnames int `json:"available_hostnames"`
	CommittedHostnames int `json:"committed_hostnames"`
	ReleasedHostnames  int `json:"released_hostnames"`
	TotalTemplates     int `json:"total_templates"`
}

// DashboardHandler handles the dashboard page
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
		BaseHandler:  *NewBaseHandler(),
		hostnameRepo: hostnameRepo,
		templateRepo: templateRepo,
	}
}

// Show displays the dashboard page
func (h *DashboardHandler) Show(c *gin.Context) {
	// Get stats
	stats, err := h.getStats(c.Request.Context())
	if err != nil {
		h.ServerError(c, err)
		return
	}

	// Get recent activity (last 10 hostnames)
	recentHostnames, err := h.getRecentActivity(c.Request.Context())
	if err != nil {
		h.ServerError(c, err)
		return
	}

	h.RenderTemplate(c, "dashboard", gin.H{
		"Title":           "Dashboard",
		"ActivePage":      "dashboard",
		"Stats":           stats,
		"RecentHostnames": recentHostnames,
	})
}

// getStats retrieves statistics for the dashboard
func (h *DashboardHandler) getStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Get template count
	templates, total, err := h.templateRepo.List(ctx, 1, 0)
	if err != nil {
		return nil, err
	}
	stats.TotalTemplates = total

	// Get hostname counts
	reservedCount, err := h.hostnameRepo.Count(ctx, 0, models.StatusReserved)
	if err != nil {
		return nil, err
	}
	stats.ReservedHostnames = reservedCount

	committedCount, err := h.hostnameRepo.Count(ctx, 0, models.StatusCommitted)
	if err != nil {
		return nil, err
	}
	stats.CommittedHostnames = committedCount

	releasedCount, err := h.hostnameRepo.Count(ctx, 0, models.StatusReleased)
	if err != nil {
		return nil, err
	}
	stats.ReleasedHostnames = releasedCount

	// Calculate totals
	stats.TotalHostnames = reservedCount + committedCount + releasedCount
	stats.AvailableHostnames = stats.TotalHostnames - reservedCount - committedCount

	return stats, nil
}

// getRecentActivity retrieves recent hostname activity
func (h *DashboardHandler) getRecentActivity(ctx context.Context) ([]*models.Hostname, error) {
	// This filter would get recently created hostnames
	filters := map[string]interface{}{}

	// Get the most recent 10 hostnames
	hostnames, _, err := h.hostnameRepo.List(ctx, 10, 0, filters)
	if err != nil {
		return nil, err
	}

	return hostnames, nil
}
