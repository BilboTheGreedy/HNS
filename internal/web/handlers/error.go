package handlers

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ErrorHandler handles error-related requests
type ErrorHandler struct {
	BaseHandler
}

// NewErrorHandler creates a new ErrorHandler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		BaseHandler: *NewBaseHandler(),
	}
}

// NotFound handles 404 errors
func (h *ErrorHandler) NotFound(c *gin.Context) {
	log.Debug().
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Str("client_ip", c.ClientIP()).
		Msg("Page not found")

	// Render 404 page
	c.HTML(http.StatusNotFound, "404.html", gin.H{
		"Title":       "Page Not Found",
		"CurrentYear": h.getCurrentYear(),
		"Path":        c.Request.URL.Path,
	})
}

// ServerError handles 500 errors
func (h *ErrorHandler) ServerError(c *gin.Context, err error) {
	// Log the error with stack trace
	log.Error().
		Err(err).
		Str("stack", string(debug.Stack())).
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Str("client_ip", c.ClientIP()).
		Msg("Server error")

	// Render 500 page
	c.HTML(http.StatusInternalServerError, "500.html", gin.H{
		"Title":       "Server Error",
		"CurrentYear": h.getCurrentYear(),
	})
}

// Forbidden handles 403 errors
func (h *ErrorHandler) Forbidden(c *gin.Context) {
	log.Debug().
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Str("client_ip", c.ClientIP()).
		Msg("Access forbidden")

	// Render 403 page
	c.HTML(http.StatusForbidden, "403.html", gin.H{
		"Title":       "Access Denied",
		"CurrentYear": h.getCurrentYear(),
		"Message":     "You do not have permission to access this page",
	})
}

// getCurrentYear returns the current year for copyright notices
func (h *ErrorHandler) getCurrentYear() int {
	return time.Now().Year()
}

// RecoveryMiddleware recovers from panics and renders a 500 page
func (h *ErrorHandler) RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch e := r.(type) {
				case error:
					err = e
				default:
					err = fmt.Errorf("%v", e)
				}

				// Log the panic
				log.Error().
					Interface("recovery", r).
					Str("stack", string(debug.Stack())).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("client_ip", c.ClientIP()).
					Msg("Recovered from panic")

				// Render 500 page
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"Title":       "Server Error",
					"CurrentYear": h.getCurrentYear(),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
