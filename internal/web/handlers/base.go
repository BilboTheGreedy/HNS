package handlers

import (
	"net/http"
	"time"

	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/gin-gonic/gin"
)

// Alert represents a flash message to show to the user
type Alert struct {
	Type    string // success, danger, warning, info
	Message string
}

// BaseHandler provides common functionality for all handlers
type BaseHandler struct{}

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
	data["CurrentYear"] = time.Now().Year()

	// Check if user is logged in and add user data to context
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

	// Set active page for navigation highlighting if not already set
	if _, ok := data["ActivePage"]; !ok {
		data["ActivePage"] = templateName
	}

	// Render template
	c.HTML(http.StatusOK, templateName+".html", data)
}

// RedirectWithAlert redirects with an alert message
func (h *BaseHandler) RedirectWithAlert(c *gin.Context, url, alertType, message string) {
	helpers.SetAlert(c, alertType, message)
	c.Redirect(http.StatusFound, url)
}

// NotFound renders a 404 page
func (h *BaseHandler) NotFound(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", gin.H{
		"Title":       "Page Not Found",
		"CurrentYear": time.Now().Year(),
	})
}

// ServerError renders a 500 page
func (h *BaseHandler) ServerError(c *gin.Context, err error) {
	c.HTML(http.StatusInternalServerError, "500.html", gin.H{
		"Title":       "Server Error",
		"CurrentYear": time.Now().Year(),
	})
}

// Forbidden renders a 403 page
func (h *BaseHandler) Forbidden(c *gin.Context) {
	c.HTML(http.StatusForbidden, "403.html", gin.H{
		"Title":       "Access Denied",
		"CurrentYear": time.Now().Year(),
	})
}
