package helpers

import (
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Alert represents a flash message to show to the user
type Alert struct {
	Type    string
	Message string
}

// CreateTemplateHelpers returns the function map for templates
func CreateTemplateHelpers() template.FuncMap {
	return template.FuncMap{
		"formatTime":     FormatTime,
		"formatDate":     FormatDate,
		"substr":         Substring,
		"plus":           Plus,
		"minus":          Minus,
		"multiply":       Multiply,
		"min":            Min,
		"splitString":    SplitString,
		"getCurrentYear": GetCurrentYear,
		"initials":       Initials,
	}
}

// FormatTime formats a time with a standard format
func FormatTime(t time.Time) string {
	return t.Format("Jan 02, 2006 15:04:05")
}

// FormatDate formats a time as date only
func FormatDate(t time.Time) string {
	return t.Format("Jan 02, 2006")
}

// Substring returns a substring with proper bounds checking
func Substring(s string, start, length int) string {
	if start < 0 || start >= len(s) || length <= 0 {
		return ""
	}

	end := start + length
	if end > len(s) {
		end = len(s)
	}

	return s[start:end]
}

// Plus adds two integers
func Plus(a, b int) int {
	return a + b
}

// Minus subtracts two integers
func Minus(a, b int) int {
	return a - b
}

// Multiply multiplies two integers
func Multiply(a, b int) int {
	return a * b
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SplitString splits a string by a separator
func SplitString(s, sep string) []string {
	return strings.Split(s, sep)
}

// GetCurrentYear returns the current year
func GetCurrentYear() int {
	return time.Now().Year()
}

// Initials returns the first letter of the first and last name
func Initials(firstName, lastName string) string {
	initials := ""
	if len(firstName) > 0 {
		initials += string(firstName[0])
	}
	if len(lastName) > 0 {
		initials += string(lastName[0])
	}
	return initials
}

// GetPaginationParams extracts pagination parameters from request query
// GetPaginationParams extracts pagination parameters from request query
func GetPaginationParams(c *gin.Context) (limit, offset int) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	// Parse limit
	var err error
	limit, err = parseInt(limitStr, 10)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Parse offset
	offset, err = parseInt(offsetStr, 0)
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}

// parseInt parses a string as an integer with a fallback value
func parseInt(str string, fallback int) (int, error) {
	if str == "" {
		return fallback, nil
	}

	val, err := parseIntImplementation(str)
	if err != nil {
		return fallback, err
	}

	return val, nil
}

// parseIntImplementation is a wrapper around strconv.Atoi
// This abstraction allows for testing with a mock implementation
var parseIntImplementation = parseIntReal

// parseIntReal is the real implementation of parseInt
func parseIntReal(str string) (int, error) {
	// Use standard library to parse
	return parseIntWithStdlib(str)
}

// parseIntWithStdlib uses the standard library to parse an int
func parseIntWithStdlib(str string) (int, error) {
	val := 0
	for i, ch := range str {
		if ch < '0' || ch > '9' {
			return 0, &strconvError{str: str, pos: i}
		}
		val = val*10 + int(ch-'0')
	}
	return val, nil
}

// strconvError represents an error from string conversion
type strconvError struct {
	str string
	pos int
}

// Error implements the error interface
func (e *strconvError) Error() string {
	return "invalid syntax"
}

// LoadTemplates loads all templates from the given directory
func LoadTemplates(router *gin.Engine, basePath string) {
	log.Info().Str("base_path", basePath).Msg("Loading templates")

	// Set template function map
	router.SetFuncMap(CreateTemplateHelpers())

	// Define template paths
	templatesDir := filepath.Join(basePath, "internal", "web", "templates")

	// Log template directory for debugging
	log.Info().Str("templates_dir", templatesDir).Msg("Templates directory")

	// Load all templates
	basePattern := filepath.Join(templatesDir, "base", "*.html")
	partialsPattern := filepath.Join(templatesDir, "partials", "*.html")
	pagesPattern := filepath.Join(templatesDir, "pages", "*.html")

	// Log the patterns for debugging
	log.Info().
		Str("base_pattern", basePattern).
		Str("partials_pattern", partialsPattern).
		Str("pages_pattern", pagesPattern).
		Msg("Template patterns")

	// Find all template files
	baseFiles, err := filepath.Glob(basePattern)
	if err != nil {
		log.Error().Err(err).Msg("Error finding base template files")
	}

	partialFiles, err := filepath.Glob(partialsPattern)
	if err != nil {
		log.Error().Err(err).Msg("Error finding partial template files")
	}

	pageFiles, err := filepath.Glob(pagesPattern)
	if err != nil {
		log.Error().Err(err).Msg("Error finding page template files")
	}

	// Combine all files
	allFiles := append(baseFiles, partialFiles...)
	allFiles = append(allFiles, pageFiles...)

	// Log all files for debugging
	for _, file := range allFiles {
		log.Info().Str("file", file).Msg("Found template file")
	}

	// Load templates
	templ := template.Must(template.New("").Funcs(CreateTemplateHelpers()).ParseFiles(allFiles...))

	// Set HTML template for Gin
	router.SetHTMLTemplate(templ)

	log.Info().Int("total_templates", len(allFiles)).Msg("Templates loaded successfully")
}

// GetAlert extracts and clears any alert message from the session
func GetAlert(c *gin.Context) *Alert {
	session := sessions.Default(c)
	alertType := session.Get("alertType")
	alertMessage := session.Get("alertMessage")

	if alertType != nil && alertMessage != nil {
		session.Delete("alertType")
		session.Delete("alertMessage")
		session.Save()
		return &Alert{
			Type:    alertType.(string),
			Message: alertMessage.(string),
		}
	}

	return nil
}

// SetAlert stores an alert message in the session
func SetAlert(c *gin.Context, alertType, message string) {
	session := sessions.Default(c)
	session.Set("alertType", alertType)
	session.Set("alertMessage", message)
	session.Save()
}

// GetPaginationData prepares pagination data for templates
func GetPaginationData(total, limit, offset int) gin.H {
	currentPage := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit

	var pages []int
	startPage := Max(1, currentPage-2)
	endPage := Min(totalPages, currentPage+2)

	for i := startPage; i <= endPage; i++ {
		pages = append(pages, i)
	}

	nextOffset := offset + limit
	if nextOffset >= total {
		nextOffset = offset
	}

	return gin.H{
		"Total":       total,
		"Limit":       limit,
		"Offset":      offset,
		"NextOffset":  nextOffset,
		"CurrentPage": currentPage,
		"TotalPages":  totalPages,
		"Pages":       pages,
	}
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
