package handlers

import (
	"net/http"
	"time"

	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
}

// NewBaseHandler creates a new BaseHandler
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{}
}

// RenderTemplate renders a template with common data
func (h *BaseHandler) RenderTemplate(c *gin.Context, templateName string, data gin.H) {
	// Initialize data map if nil
	if data == nil {
		data = gin.H{}
	}

	// Add common data
	loggedIn, exists := c.Get("loggedIn")
	if exists && loggedIn.(bool) {
		data["LoggedIn"] = true
		data["Username"], _ = c.Get("username")
		data["IsAdmin"], _ = c.Get("isAdmin")
	} else {
		data["LoggedIn"] = false
	}

	// Add alert message if available
	alert := helpers.GetAlert(c)
	if alert != nil {
		data["Alert"] = alert
	}

	// Add current year for footer
	data["CurrentYear"] = time.Now().Year()

	// Set active page for navigation highlighting
	if _, ok := data["ActivePage"]; !ok {
		data["ActivePage"] = templateName
	}

	// Log template rendering
	log.Debug().
		Str("template", templateName).
		Interface("data_keys", getMapKeys(data)).
		Msg("Rendering template")

	// Special case for login and register pages which are full templates
	if templateName == "login" || templateName == "register" {
		c.HTML(http.StatusOK, "pages/"+templateName+".html", data)
		return
	}

	// Otherwise render with base layout
	c.HTML(http.StatusOK, "base.html", data)
}

// getMapKeys returns the keys of a map for logging
func getMapKeys(m gin.H) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// RedirectWithAlert redirects with an alert message
func (h *BaseHandler) RedirectWithAlert(c *gin.Context, url, alertType, message string) {
	helpers.SetAlert(c, alertType, message)
	c.Redirect(http.StatusFound, url)
}

// NotFound renders a 404 page
func (h *BaseHandler) NotFound(c *gin.Context) {
	c.HTML(http.StatusNotFound, "pages/404.html", gin.H{
		"Title":       "Page Not Found",
		"CurrentYear": time.Now().Year(),
	})
}

// Forbidden renders a 403 page
func (h *BaseHandler) Forbidden(c *gin.Context) {
	c.HTML(http.StatusForbidden, "pages/403.html", gin.H{
		"Title":       "Access Denied",
		"CurrentYear": time.Now().Year(),
	})
}

// InternalError renders a 500 page
func (h *BaseHandler) InternalError(c *gin.Context, err error) {
	if err != nil {
		log.Error().Err(err).Str("path", c.Request.URL.Path).Msg("Internal server error")
	}

	c.HTML(http.StatusInternalServerError, "pages/500.html", gin.H{
		"Title":       "Internal Server Error",
		"CurrentYear": time.Now().Year(),
	})
}

// Home renders the home/dashboard page
func (h *BaseHandler) Home(c *gin.Context) {
	// If not logged in, redirect to login page
	loggedIn, exists := c.Get("loggedIn")
	if !exists || !loggedIn.(bool) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// For logged-in users, redirect to dashboard
	c.Redirect(http.StatusFound, "/dashboard")
}
