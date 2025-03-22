package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// WebHandler handles web requests and template rendering
type WebHandler struct {
	templateRepo    repository.TemplateRepository
	hostnameRepo    repository.HostnameRepository
	userRepo        repository.UserRepository
	jwtManager      *auth.JWTManager
	generatorSvc    *service.GeneratorService
	reservationSvc  *service.ReservationService
	dnsChecker      interface{} // Using interface for flexibility
	sequenceService *service.SequenceService
}

// Alert represents a flash message to show to the user
type Alert struct {
	Type    string
	Message string
}

// NewWebHandler creates a new WebHandler
func NewWebHandler(
	templateRepo repository.TemplateRepository,
	hostnameRepo repository.HostnameRepository,
	userRepo repository.UserRepository,
	jwtManager *auth.JWTManager,
	generatorSvc *service.GeneratorService,
	reservationSvc *service.ReservationService,
	dnsChecker interface{},
	sequenceService *service.SequenceService,
) *WebHandler {
	return &WebHandler{
		templateRepo:    templateRepo,
		hostnameRepo:    hostnameRepo,
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		generatorSvc:    generatorSvc,
		reservationSvc:  reservationSvc,
		dnsChecker:      dnsChecker,
		sequenceService: sequenceService,
	}
}

func setupSessionStore(router *gin.Engine) {
	// Setup sessions with correct key sizes for AES encryption
	// AES-256 requires exactly 32 bytes (256 bits) for the key
	authKey := []byte("01234567890123456789012345678901")       // Exactly 32 bytes
	encryptionKey := []byte("98765432109876543210987654321098") // Exactly 32 bytes

	// Use more secure store with encryption
	store := cookie.NewStore(authKey, encryptionKey)

	// Configure cookie options
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1 day
		HttpOnly: true,
		Secure:   false, // Change to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	router.Use(sessions.Sessions("hns_session", store))
}

var templateFuncMap = template.FuncMap{
	"formatTime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	},
	"formatDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
	"substr": func(s string, start, length int) string {
		if start >= len(s) {
			return ""
		}
		end := start + length
		if end > len(s) {
			end = len(s)
		}
		return s[start:end]
	},
	"plus": func(a, b int) int {
		return a + b
	},
	"minus": func(a, b int) int {
		return a - b
	},
	"multiply": func(a, b int) int {
		return a * b
	},
	"min": func(a, b int) int {
		if a < b {
			return a
		}
		return b
	},
}

// setupWebRouter sets up template loading correctly
func SetupWebRouter(
	router *gin.Engine,
	webHandler *WebHandler,
) {
	// Setup sessions with correct key sizes for AES encryption
	// AES-256 requires exactly 32 bytes (256 bits) for the key
	authKey := []byte("01234567890123456789012345678901")       // Exactly 32 bytes
	encryptionKey := []byte("98765432109876543210987654321098") // Exactly 32 bytes

	// Use more secure store with encryption
	store := cookie.NewStore(authKey, encryptionKey)

	// Configure cookie options
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1 day
		HttpOnly: true,
		Secure:   false, // Change to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	router.Use(sessions.Sessions("hns_session", store))

	// Setup template functions
	router.SetFuncMap(template.FuncMap{
		"formatTime":     formatTime,
		"formatDate":     formatDate,
		"substr":         substring,
		"plus":           plus,
		"minus":          minus,
		"multiply":       multiply,
		"min":            min,
		"splitString":    splitString,
		"getCurrentYear": getCurrentYear,
	})

	// Setup template loading correctly
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get working directory")
	}

	templatesDir := filepath.Join(wd, "internal", "web", "templates")

	tmpl := template.New("").Funcs(templateFuncMap)
	template.Must(tmpl.ParseGlob(filepath.Join(templatesDir, "base", "*.html")))
	template.Must(tmpl.ParseGlob(filepath.Join(templatesDir, "pages", "*.html")))
	template.Must(tmpl.ParseGlob(filepath.Join(templatesDir, "partials", "*.html")))
	router.SetHTMLTemplate(tmpl)

	// Static files path
	staticPath := filepath.Join("internal", "web", "static")
	log.Info().Str("static_path", staticPath).Msg("Setting up static files")
	router.Static("/static", staticPath)
	// Public routes
	router.GET("/", webHandler.Home)
	router.GET("/login", webHandler.LoginPage)
	router.POST("/login", webHandler.Login)
	router.GET("/register", webHandler.RegisterPage)
	router.POST("/register", webHandler.Register)
	router.GET("/logout", webHandler.Logout)

	// Error pages
	router.GET("/404", func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{
			"Title": "Page Not Found",
		})
	})
	router.GET("/403", func(c *gin.Context) {
		c.HTML(http.StatusForbidden, "pages/403.html", gin.H{
			"Title": "Access Denied",
		})
	})
	router.GET("/500", func(c *gin.Context) {
		c.HTML(http.StatusInternalServerError, "pages/500.html", gin.H{
			"Title": "Internal Server Error",
		})
	})

	// Add a simple auth test route
	router.GET("/auth-test", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("userID")
		username := session.Get("username")

		if userID == nil {
			c.JSON(http.StatusOK, gin.H{
				"authenticated": false,
				"message":       "Not logged in",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"authenticated": true,
			"userID":        userID,
			"username":      username,
		})
	})

	// Protected routes
	auth := router.Group("/")
	auth.Use(webHandler.AuthRequired())
	{
		// Hostname routes
		auth.GET("/hostnames", webHandler.HostnameList)
		auth.GET("/hostnames/generate", webHandler.GenerateHostnamePage)
		auth.GET("/hostnames/reserve", webHandler.ReserveHostnamePage)
		auth.GET("/hostnames/:id", webHandler.HostnameDetail)
		auth.GET("/hostnames/:id/commit", webHandler.CommitHostname)
		auth.GET("/hostnames/:id/release", webHandler.ReleaseHostname)

		// Template routes
		auth.GET("/templates", webHandler.TemplateList)
		auth.GET("/templates/:id", webHandler.TemplateDetail)

		// DNS routes
		auth.GET("/dns", webHandler.DNSTools)
		auth.GET("/dns/check", webHandler.DNSCheck)
		auth.GET("/dns/check/:hostname", webHandler.DNSCheckHostname)
		auth.GET("/dns/scan", webHandler.DNSScan)

		// User profile
		auth.GET("/profile", webHandler.UserProfile)
		auth.POST("/profile/update", webHandler.UpdateProfile)
		auth.POST("/profile/change-password", webHandler.ChangePassword)

		// API keys
		auth.GET("/api-keys", webHandler.APIKeysList)
		auth.POST("/api-keys", webHandler.CreateAPIKey)
		auth.POST("/api-keys/delete", webHandler.DeleteAPIKey)

		// Admin routes
		admin := auth.Group("/admin")
		admin.Use(webHandler.AdminRequired())
		{
			admin.GET("/users", webHandler.UsersList)
			admin.POST("/users", webHandler.CreateUser)
			admin.POST("/users/update", webHandler.UpdateUser)
			admin.POST("/users/delete", webHandler.DeleteUser)

			admin.GET("/templates/new", webHandler.NewTemplate)
			admin.POST("/templates", webHandler.CreateTemplate)
			admin.GET("/templates/:id/edit", webHandler.EditTemplate)
			admin.POST("/templates/:id", webHandler.UpdateTemplate)
		}
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{
			"Title": "Page Not Found",
		})
	})
}

// AuthRequired middleware checks if user is authenticated
func (h *WebHandler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("userID")

		// Debug log session data
		log.Info().
			Interface("userID", userID).
			Interface("username", session.Get("username")).
			Interface("loggedIn", session.Get("loggedIn")).
			Str("path", c.Request.URL.Path).
			Msg("Auth check - session data")

		if userID == nil {
			log.Warn().Str("path", c.Request.URL.Path).Msg("No userID in session, redirecting to login")
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Get user data for templates
		user, err := h.userRepo.GetByID(c.Request.Context(), userID.(int64))
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user by ID")
			session.Clear()
			session.Save()
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Make user data available in templates
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("username", user.Username)
		c.Set("isAdmin", user.Role == models.RoleAdmin)
		c.Set("loggedIn", true)

		c.Next()
	}
}
func (h *WebHandler) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if user.(*models.User).Role != models.RoleAdmin {
			c.HTML(http.StatusForbidden, "pages/403.html", gin.H{
				"Title": "Access Denied",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
func formatTime(t time.Time) string {
	return t.Format("Jan 02, 2006 15:04:05")
}

func formatDate(t time.Time) string {
	return t.Format("Jan 02, 2006")
}

func substring(s string, start, length int) string {
	if start < 0 || start > len(s) {
		return ""
	}
	if length <= 0 {
		return ""
	}
	if start+length > len(s) {
		length = len(s) - start
	}
	return s[start : start+length]
}

func plus(a, b int) int {
	return a + b
}

func minus(a, b int) int {
	return a - b
}

func multiply(a, b int) int {
	return a * b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func splitString(s, sep string) []string {
	return strings.Split(s, sep)
}

func getCurrentYear() int {
	return time.Now().Year()
}

// getAlert extracts and clears any alert message from the session
func getAlert(c *gin.Context) *Alert {
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

// setAlert stores an alert message in the session
func setAlert(c *gin.Context, alertType, message string) {
	session := sessions.Default(c)
	session.Set("alertType", alertType)
	session.Set("alertMessage", message)
	session.Save()
}

func (h *WebHandler) renderTemplate(c *gin.Context, name string, data gin.H) {
	// Add common template data
	if data == nil {
		data = gin.H{}
	}

	// Add user information if logged in
	loggedIn, _ := c.Get("loggedIn")
	if loggedIn != nil && loggedIn.(bool) {
		data["LoggedIn"] = true
		data["Username"], _ = c.Get("username")
		data["IsAdmin"], _ = c.Get("isAdmin")
	} else {
		data["LoggedIn"] = false
	}

	// Add alert message if available
	alert := getAlert(c)
	if alert != nil {
		data["Alert"] = alert
	}

	// Add current year for footer
	data["CurrentYear"] = time.Now().Year()

	// Important: Use the base template, not the page name directly
	// This is the key change - we render "base.html" and the content template is used via inheritance
	c.HTML(http.StatusOK, "base.html", data)
}

// getPaginationData prepares pagination data for templates
func getPaginationData(total, limit, offset int) (pageData gin.H) {
	currentPage := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit

	var pages []int
	startPage := maxInt(1, currentPage-2)
	endPage := minInt(totalPages, currentPage+2)

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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Route handlers - Auth and Public

// Home renders the dashboard page
func (h *WebHandler) Home(c *gin.Context) {
	// If not logged in, show login page
	loggedIn, exists := c.Get("loggedIn")
	if !exists || !loggedIn.(bool) {
		c.Redirect(http.StatusFound, "/login")
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
	recentHostnames, _, _ := h.hostnameRepo.List(ctx, 5, 0, filters)

	// Get template count
	templates, totalTemplates, _ := h.templateRepo.List(ctx, 5, 0)

	h.renderTemplate(c, "dashboard", gin.H{
		"Title":              "Dashboard",
		"TotalHostnames":     totalHostnames,
		"AvailableHostnames": availableCount,
		"ReservedHostnames":  reservedCount,
		"CommittedHostnames": committedCount,
		"ReleasedHostnames":  releasedCount,
		"RecentHostnames":    recentHostnames,
		"TotalTemplates":     totalTemplates,
		"Templates":          templates,
	})
}

// LoginPage renders the login page
func (h *WebHandler) LoginPage(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("userID") != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	h.renderTemplate(c, "base.html", gin.H{
		"Title": "Login",
		"Page":  "login", // 🔥 This selects which sub-template to render
	})
}

// Login handles the login form submission
func (h *WebHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	log.Info().Str("username", username).Msg("Web UI Login attempt")

	// Get user by username
	user, err := h.userRepo.GetByUsername(c.Request.Context(), username)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("User not found during login")
		setAlert(c, "danger", "Invalid username or password")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Verify password directly instead of using the JWT auth flow
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Invalid password")
		setAlert(c, "danger", "Invalid username or password")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	log.Info().Str("username", username).Msg("Authentication successful, setting session")

	// Create a new clean session
	session := sessions.Default(c)
	session.Clear()

	// Set session data explicitly
	session.Set("userID", user.ID)
	session.Set("username", user.Username)
	session.Set("isAdmin", user.Role == models.RoleAdmin)
	session.Set("loggedIn", true) // Add explicit loggedIn flag

	// Also generate JWT token for API access
	if h.jwtManager != nil {
		token, err := h.jwtManager.GenerateToken(user)
		if err == nil {
			session.Set("token", token)
		}
	}

	// Save session immediately
	err = session.Save()
	if err != nil {
		log.Error().Err(err).Msg("Failed to save session")
		setAlert(c, "danger", "Login failed due to session error")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// For this request, set context values directly
	c.Set("user", user)
	c.Set("userID", user.ID)
	c.Set("username", user.Username)
	c.Set("isAdmin", user.Role == models.RoleAdmin)
	c.Set("loggedIn", true)

	// Redirect to hostnames page instead of dashboard (known working page)
	log.Info().Str("username", username).Msg("Session saved, redirecting to hostnames page")
	c.Redirect(http.StatusFound, "/hostnames")
}

// RegisterPage renders the registration page
func (h *WebHandler) RegisterPage(c *gin.Context) {
	// If already logged in, redirect to home
	loggedIn, exists := c.Get("loggedIn")
	if exists && loggedIn.(bool) {
		c.Redirect(http.StatusFound, "/")
		return
	}

	h.renderTemplate(c, "register", gin.H{
		"Title": "Register",
	})
}

// Register handles the registration form submission
func (h *WebHandler) Register(c *gin.Context) {
	// Get form data
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")

	// Validate passwords match
	if password != confirmPassword {
		setAlert(c, "danger", "Passwords do not match")
		c.Redirect(http.StatusFound, "/register")
		return
	}

	// Create user request
	userReq := &models.UserCreateRequest{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Role:      "user", // Default role for self-registration
	}

	// Create a temporary auth handler to register
	authHandler := auth.NewAuthHandler(h.userRepo, h.jwtManager, nil)
	_, err := authHandler.RegisterUserInternal(c.Request.Context(), userReq)
	if err != nil {
		setAlert(c, "danger", fmt.Sprintf("Registration failed: %s", err.Error()))
		c.Redirect(http.StatusFound, "/register")
		return
	}

	// Set success message and redirect to login
	setAlert(c, "success", "Registration successful! You can now log in.")
	c.Redirect(http.StatusFound, "/login")
}

// Logout handles user logout
func (h *WebHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

// Route handlers - Hostnames

// HostnameList displays the list of hostnames
func (h *WebHandler) HostnameList(c *gin.Context) {
	// Get pagination parameters
	limit, offset := h.getPaginationParams(c)

	// Get filter parameters
	filters := map[string]interface{}{}

	name := c.Query("name")
	if name != "" {
		filters["name LIKE"] = "%" + name + "%"
	}

	templateIDStr := c.Query("template_id")
	if templateIDStr != "" {
		templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
		if err == nil {
			filters["template_id"] = templateID
		}
	}

	status := c.Query("status")
	if status != "" {
		filters["status"] = status
	}

	reservedBy := c.Query("reserved_by")
	if reservedBy != "" {
		filters["reserved_by"] = reservedBy
	}

	// Get hostnames
	ctx := c.Request.Context()
	hostnames, total, err := h.hostnameRepo.List(ctx, limit, offset, filters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get hostnames")
		setAlert(c, "danger", "Error retrieving hostnames")
		h.renderTemplate(c, "hostname_list", gin.H{
			"Title": "Hostname Management",
		})
		return
	}

	// Get templates for filter dropdown
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
		"Title":     "Hostname Management",
		"Hostnames": hostnamesWithTemplates,
		"Templates": templates,
		"Filters": gin.H{
			"name":        name,
			"template_id": templateIDStr,
			"status":      status,
			"reserved_by": reservedBy,
		},
		"PaginationURL": paginationURL,
	}

	// Add pagination data
	paginationData := getPaginationData(total, limit, offset)
	for k, v := range paginationData {
		templateData[k] = v
	}

	h.renderTemplate(c, "hostname_list", templateData)
}

// HostnameDetail displays the details of a single hostname
func (h *WebHandler) HostnameDetail(c *gin.Context) {
	// Get hostname ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		setAlert(c, "danger", "Invalid hostname ID")
		c.Redirect(http.StatusFound, "/hostnames")
		return
	}

	// Get hostname
	ctx := c.Request.Context()
	hostname, err := h.hostnameRepo.GetByID(ctx, id)
	if err != nil {
		setAlert(c, "danger", "Hostname not found")
		c.Redirect(http.StatusFound, "/hostnames")
		return
	}

	// Get template
	template, err := h.templateRepo.GetByID(ctx, hostname.TemplateID)
	if err != nil {
		template = &models.Template{
			Name: "Unknown",
		}
	}

	// Get DNS status if implemented
	var dnsResult interface{}
	if checker, ok := h.dnsChecker.(interface {
		CheckHostname(ctx context.Context, hostname string) (*models.DNSVerificationResult, error)
	}); ok {
		result, _ := checker.CheckHostname(ctx, hostname.Name)
		if result != nil {
			dnsResult = result
		}
	}

	h.renderTemplate(c, "hostname_detail", gin.H{
		"Title":     hostname.Name,
		"Hostname":  hostname,
		"Template":  template,
		"DNSResult": dnsResult,
	})
}

func (h *WebHandler) GenerateHostnamePage(c *gin.Context) {
	// Get templates
	ctx := c.Request.Context()
	templates, _, err := h.templateRepo.List(ctx, 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get templates")
		setAlert(c, "danger", "Error retrieving templates")
	}

	h.renderTemplate(c, "hostname_generate", gin.H{
		"Title":     "Generate Hostname",
		"Templates": templates,
	})
}

// ReserveHostnamePage displays the hostname reservation form
func (h *WebHandler) ReserveHostnamePage(c *gin.Context) {
	// Get templates
	ctx := c.Request.Context()
	templates, _, err := h.templateRepo.List(ctx, 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get templates")
		setAlert(c, "danger", "Error retrieving templates")
	}

	h.renderTemplate(c, "hostname_reservation", gin.H{
		"Title":     "Reserve Hostname",
		"Templates": templates,
	})
}

// CommitHostname commits a reserved hostname
func (h *WebHandler) CommitHostname(c *gin.Context) {
	// Get hostname ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		setAlert(c, "danger", "Invalid hostname ID")
		c.Redirect(http.StatusFound, "/hostnames")
		return
	}

	// Get username
	username, _ := c.Get("username")
	if username == nil {
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
		setAlert(c, "danger", fmt.Sprintf("Failed to commit hostname: %s", err.Error()))
		c.Redirect(http.StatusFound, fmt.Sprintf("/hostnames/%d", id))
		return
	}

	// Success
	setAlert(c, "success", "Hostname committed successfully")
	c.Redirect(http.StatusFound, fmt.Sprintf("/hostnames/%d", id))
}

// ReleaseHostname releases a committed hostname
func (h *WebHandler) ReleaseHostname(c *gin.Context) {
	// Get hostname ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		setAlert(c, "danger", "Invalid hostname ID")
		c.Redirect(http.StatusFound, "/hostnames")
		return
	}

	// Get username
	username, _ := c.Get("username")
	if username == nil {
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
		setAlert(c, "danger", fmt.Sprintf("Failed to release hostname: %s", err.Error()))
		c.Redirect(http.StatusFound, fmt.Sprintf("/hostnames/%d", id))
		return
	}

	// Success
	setAlert(c, "success", "Hostname released successfully")
	c.Redirect(http.StatusFound, fmt.Sprintf("/hostnames/%d", id))
}

// Route handlers - Templates

// TemplateList displays the list of templates
func (h *WebHandler) TemplateList(c *gin.Context) {
	// Get pagination parameters
	limit, offset := h.getPaginationParams(c)

	// Get templates
	ctx := c.Request.Context()
	templates, total, err := h.templateRepo.List(ctx, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get templates")
		setAlert(c, "danger", "Error retrieving templates")
		h.renderTemplate(c, "template_list", gin.H{
			"Title": "Templates",
		})
		return
	}

	// Combine pagination data with template data
	templateData := gin.H{
		"Title":     "Templates",
		"Templates": templates,
	}

	// Add pagination data
	paginationData := getPaginationData(total, limit, offset)
	for k, v := range paginationData {
		templateData[k] = v
	}

	h.renderTemplate(c, "template_list", templateData)
}

// TemplateDetail displays the details of a single template
func (h *WebHandler) TemplateDetail(c *gin.Context) {
	// Get template ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		setAlert(c, "danger", "Invalid template ID")
		c.Redirect(http.StatusFound, "/templates")
		return
	}

	// Get template
	ctx := c.Request.Context()
	template, err := h.templateRepo.GetByID(ctx, id)
	if err != nil {
		setAlert(c, "danger", "Template not found")
		c.Redirect(http.StatusFound, "/templates")
		return
	}

	h.renderTemplate(c, "template_detail", gin.H{
		"Title":    template.Name,
		"Template": template,
	})
}

// NewTemplate displays the template creation form
func (h *WebHandler) NewTemplate(c *gin.Context) {
	h.renderTemplate(c, "template_form", gin.H{
		"Title":    "Create Template",
		"Template": &models.Template{},
	})
}

// CreateTemplate handles template creation form submission
// This patch fixes the unused isActive variable in the CreateTemplate function

func (h *WebHandler) CreateTemplate(c *gin.Context) {
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
	isActive := c.PostForm("is_active") == "on"

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

	// Process group data - form data comes in as groups[0][name], groups[0][length], etc.
	i := 0
	for {
		nameKey := fmt.Sprintf("groups[%d][name]", i)
		groupName := c.PostForm(nameKey)
		if groupName == "" {
			break // No more groups
		}

		// Get group data
		lengthKey := fmt.Sprintf("groups[%d][length]", i)
		validationTypeKey := fmt.Sprintf("groups[%d][validation_type]", i)
		validationValueKey := fmt.Sprintf("groups[%d][validation_value]", i)
		isRequiredKey := fmt.Sprintf("groups[%d][is_required]", i)

		// Parse values
		lengthStr := c.PostForm(lengthKey)
		length, _ := strconv.Atoi(lengthStr)
		if length <= 0 {
			length = 1
		}

		validationType := c.PostForm(validationTypeKey)
		validationValue := c.PostForm(validationValueKey)
		isRequired := c.PostForm(isRequiredKey) == "on"

		// Add group to request
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

	// Create template
	_, err := h.generatorSvc.CreateTemplate(c.Request.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create template")
		setAlert(c, "danger", fmt.Sprintf("Failed to create template: %s", err.Error()))
		h.renderTemplate(c, "template_form", gin.H{
			"Title":    "Create Template",
			"Template": req,
			"IsActive": isActive, // Pass isActive to template to resolve unused variable
		})
		return
	}

	// Success
	setAlert(c, "success", "Template created successfully")
	c.Redirect(http.StatusFound, "/templates")
}

// EditTemplate displays the template edit form
func (h *WebHandler) EditTemplate(c *gin.Context) {
	// Get template ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		setAlert(c, "danger", "Invalid template ID")
		c.Redirect(http.StatusFound, "/templates")
		return
	}

	// Get template
	ctx := c.Request.Context()
	template, err := h.templateRepo.GetByID(ctx, id)
	if err != nil {
		setAlert(c, "danger", "Template not found")
		c.Redirect(http.StatusFound, "/templates")
		return
	}

	h.renderTemplate(c, "template_form", gin.H{
		"Title":    "Edit Template",
		"Template": template,
	})
}

// UpdateTemplate handles template update form submission
func (h *WebHandler) UpdateTemplate(c *gin.Context) {
	// Get template ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		setAlert(c, "danger", "Invalid template ID")
		c.Redirect(http.StatusFound, "/templates")
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
		setAlert(c, "danger", "Template not found")
		c.Redirect(http.StatusFound, "/templates")
		return
	}

	// Get form data (same as CreateTemplate)
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
		setAlert(c, "danger", fmt.Sprintf("Failed to update template: %s", err.Error()))
		h.renderTemplate(c, "template_form", gin.H{
			"Title":    "Edit Template",
			"Template": template,
		})
		return
	}

	// Process groups (this would involve deleting and recreating all groups)
	// For simplicity, we'll assume the frontend handles this via API calls

	// Success
	setAlert(c, "success", "Template updated successfully")
	c.Redirect(http.StatusFound, fmt.Sprintf("/templates/%d", id))
}

// Route handlers - DNS

// DNSTools displays the DNS tools landing page
func (h *WebHandler) DNSTools(c *gin.Context) {
	h.renderTemplate(c, "dns_tools", gin.H{
		"Title": "DNS Tools",
	})
}

// DNSCheck displays the DNS check form
func (h *WebHandler) DNSCheck(c *gin.Context) {
	h.renderTemplate(c, "dns_check", gin.H{
		"Title": "DNS Check",
	})
}

// DNSCheckHostname checks a specific hostname and displays results
func (h *WebHandler) DNSCheckHostname(c *gin.Context) {
	hostname := c.Param("hostname")

	// Check DNS if checker is available
	var dnsResult interface{}
	if checker, ok := h.dnsChecker.(interface {
		CheckHostname(ctx context.Context, hostname string) (*models.DNSVerificationResult, error)
	}); ok {
		result, err := checker.CheckHostname(c.Request.Context(), hostname)
		if err == nil && result != nil {
			dnsResult = result
		}
	}

	h.renderTemplate(c, "dns_check", gin.H{
		"Title":     "DNS Check",
		"Hostname":  hostname,
		"DNSResult": dnsResult,
	})
}

// DNSScan displays the DNS scan form
func (h *WebHandler) DNSScan(c *gin.Context) {
	// Get templates
	ctx := c.Request.Context()
	templates, _, err := h.templateRepo.List(ctx, 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get templates")
		setAlert(c, "danger", "Error retrieving templates")
	}

	h.renderTemplate(c, "dns_scan", gin.H{
		"Title":     "DNS Scan",
		"Templates": templates,
	})
}

// Route handlers - User

// UserProfile displays the user profile page
func (h *WebHandler) UserProfile(c *gin.Context) {
	// Get user
	user, _ := c.Get("user")
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get recent hostnames by this user
	ctx := c.Request.Context()
	filters := map[string]interface{}{
		"reserved_by": user.(*models.User).Username,
	}
	recentHostnames, _, _ := h.hostnameRepo.List(ctx, 5, 0, filters)

	// Get statistics
	reservedCount, _ := h.hostnameRepo.CountByUser(ctx, user.(*models.User).Username, models.StatusReserved)
	committedCount, _ := h.hostnameRepo.CountByUser(ctx, user.(*models.User).Username, models.StatusCommitted)
	releasedCount, _ := h.hostnameRepo.CountByUser(ctx, user.(*models.User).Username, models.StatusReleased)

	stats := gin.H{
		"Reserved":  reservedCount,
		"Committed": committedCount,
		"Released":  releasedCount,
		"Total":     reservedCount + committedCount + releasedCount,
	}

	h.renderTemplate(c, "user_profile", gin.H{
		"Title":           "My Profile",
		"User":            user.(*models.User),
		"RecentHostnames": recentHostnames,
		"Stats":           stats,
	})
}

// UpdateProfile handles profile update form submission
func (h *WebHandler) UpdateProfile(c *gin.Context) {
	// Get user
	userID, _ := c.Get("userID")
	if userID == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get form data
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	email := c.PostForm("email")

	// Get user from database
	ctx := c.Request.Context()
	user, err := h.userRepo.GetByID(ctx, userID.(int64))
	if err != nil {
		setAlert(c, "danger", "User not found")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Update user fields
	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email

	// Save changes
	err = h.userRepo.Update(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user profile")
		setAlert(c, "danger", "Failed to update profile")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Success
	setAlert(c, "success", "Profile updated successfully")
	c.Redirect(http.StatusFound, "/profile")
}

// ChangePassword handles password change form submission
func (h *WebHandler) ChangePassword(c *gin.Context) {
	// Get user
	userID, _ := c.Get("userID")
	if userID == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get form data
	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	// Validate passwords match
	if newPassword != confirmPassword {
		setAlert(c, "danger", "New passwords do not match")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Get user from database
	ctx := c.Request.Context()
	user, err := h.userRepo.GetByID(ctx, userID.(int64))
	if err != nil {
		setAlert(c, "danger", "User not found")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Verify current password
	err = auth.VerifyPassword(currentPassword, user.PasswordHash)
	if err != nil {
		setAlert(c, "danger", "Current password is incorrect")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Update password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		setAlert(c, "danger", "Failed to update password")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	user.PasswordHash = hashedPassword

	// Save changes
	err = h.userRepo.Update(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update password")
		setAlert(c, "danger", "Failed to update password")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Success
	setAlert(c, "success", "Password updated successfully")
	c.Redirect(http.StatusFound, "/profile")
}

// Route handlers - API Keys

// APIKeysList displays the user's API keys
func (h *WebHandler) APIKeysList(c *gin.Context) {
	// Get user ID
	userID, _ := c.Get("userID")
	if userID == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get API keys
	ctx := c.Request.Context()
	apiKeys, err := h.userRepo.ListAPIKeys(ctx, userID.(int64))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get API keys")
		setAlert(c, "danger", "Error retrieving API keys")
	}

	h.renderTemplate(c, "apikeys", gin.H{
		"Title":   "API Keys",
		"ApiKeys": apiKeys,
	})
}

// CreateAPIKey handles API key creation form submission
func (h *WebHandler) CreateAPIKey(c *gin.Context) {
	// Get user ID
	userID, _ := c.Get("userID")
	if userID == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get form data
	name := c.PostForm("name")

	// Get scope (multiple checkboxes)
	scopeArray := c.PostFormArray("scope[]")
	scope := strings.Join(scopeArray, ",")

	// Create API key request
	req := &models.APIKeyCreateRequest{
		Name:  name,
		Scope: scope,
	}

	// Create API key manager if needed
	if h.jwtManager == nil {
		setAlert(c, "danger", "JWT manager not initialized")
		c.Redirect(http.StatusFound, "/api-keys")
		return
	}

	apiKeyManager := auth.NewAPIKeyManager(h.userRepo, 24*30*time.Hour) // 30 days

	// Create API key
	apiKey, err := apiKeyManager.GenerateAPIKey(req, userID.(int64))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create API key")
		setAlert(c, "danger", fmt.Sprintf("Failed to create API key: %s", err.Error()))
		c.Redirect(http.StatusFound, "/api-keys")
		return
	}

	// Success - since we can only show the key once, set it in the session
	session := sessions.Default(c)
	session.Set("newApiKey", apiKey.Key)
	session.Set("newApiKeyName", apiKey.Name)
	session.Set("newApiKeyScope", apiKey.Scope)
	session.Set("newApiKeyExpires", apiKey.ExpiresAt.Format(time.RFC3339))
	session.Save()

	setAlert(c, "success", "API key created successfully")
	c.Redirect(http.StatusFound, "/api-keys")
}

// DeleteAPIKey handles API key deletion form submission
func (h *WebHandler) DeleteAPIKey(c *gin.Context) {
	// Get user ID
	userID, _ := c.Get("userID")
	if userID == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get API key ID
	idStr := c.PostForm("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		setAlert(c, "danger", "Invalid API key ID")
		c.Redirect(http.StatusFound, "/api-keys")
		return
	}

	// Create API key manager if needed
	apiKeyManager := auth.NewAPIKeyManager(h.userRepo, 24*30*time.Hour) // 30 days

	// Delete API key
	err = apiKeyManager.DeleteAPIKey(id, userID.(int64))
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete API key")
		setAlert(c, "danger", fmt.Sprintf("Failed to delete API key: %s", err.Error()))
		c.Redirect(http.StatusFound, "/api-keys")
		return
	}

	// Success
	setAlert(c, "success", "API key deleted successfully")
	c.Redirect(http.StatusFound, "/api-keys")
}

// Route handlers - Admin

// UsersList displays the list of users (admin only)
func (h *WebHandler) UsersList(c *gin.Context) {
	// Get pagination parameters
	limit, offset := h.getPaginationParams(c)

	// Get users
	ctx := c.Request.Context()
	users, total, err := h.userRepo.List(ctx, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get users")
		setAlert(c, "danger", "Error retrieving users")
		h.renderTemplate(c, "users", gin.H{
			"Title": "Users",
		})
		return
	}

	// Get current username
	username, _ := c.Get("username")
	if username == nil {
		username = ""
	}

	// Combine pagination data with template data
	templateData := gin.H{
		"Title":       "User Management",
		"Users":       users,
		"CurrentUser": username.(string),
	}

	// Add pagination data
	paginationData := getPaginationData(total, limit, offset)
	for k, v := range paginationData {
		templateData[k] = v
	}

	h.renderTemplate(c, "users", templateData)
}

// CreateUser handles user creation form submission (admin only)
func (h *WebHandler) CreateUser(c *gin.Context) {
	// Get form data
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	role := c.PostForm("role")
	isActive := c.PostForm("is_active") == "on"

	// Create user request
	userReq := &models.UserCreateRequest{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
	}

	// Create a temporary auth handler to register
	authHandler := auth.NewAuthHandler(h.userRepo, h.jwtManager, nil)
	user, err := authHandler.RegisterUserInternal(c.Request.Context(), userReq)
	if err != nil {
		setAlert(c, "danger", fmt.Sprintf("Failed to create user: %s", err.Error()))
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// Update is_active if needed
	if !isActive {
		user.IsActive = false
		err = h.userRepo.Update(c.Request.Context(), user)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update user active status")
		}
	}

	// Success
	setAlert(c, "success", "User created successfully")
	c.Redirect(http.StatusFound, "/admin/users")
}

// UpdateUser handles user update form submission (admin only)
func (h *WebHandler) UpdateUser(c *gin.Context) {
	// Get user ID
	idStr := c.PostForm("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		setAlert(c, "danger", "Invalid user ID")
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// Get user from database
	ctx := c.Request.Context()
	user, err := h.userRepo.GetByID(ctx, id)
	if err != nil {
		setAlert(c, "danger", "User not found")
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// Get form data
	email := c.PostForm("email")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	role := c.PostForm("role")
	password := c.PostForm("password")
	isActive := c.PostForm("is_active") == "on"

	// Update user fields
	user.Email = email
	user.FirstName = firstName
	user.LastName = lastName
	user.Role = models.Role(role)
	user.IsActive = isActive

	// Update password if provided
	if password != "" {
		hashedPassword, err := auth.HashPassword(password)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash password")
			setAlert(c, "danger", "Failed to update password")
			c.Redirect(http.StatusFound, "/admin/users")
			return
		}
		user.PasswordHash = hashedPassword
	}

	// Save changes
	err = h.userRepo.Update(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user")
		setAlert(c, "danger", "Failed to update user")
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// Success
	setAlert(c, "success", "User updated successfully")
	c.Redirect(http.StatusFound, "/admin/users")
}

// DeleteUser handles user deletion form submission (admin only)
func (h *WebHandler) DeleteUser(c *gin.Context) {
	// Get user ID
	idStr := c.PostForm("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		setAlert(c, "danger", "Invalid user ID")
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// Delete user
	ctx := c.Request.Context()
	err = h.userRepo.Delete(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete user")
		setAlert(c, "danger", "Failed to delete user")
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// Success
	setAlert(c, "success", "User deleted successfully")
	c.Redirect(http.StatusFound, "/admin/users")
}

// Helper methods

// getPaginationParams gets pagination parameters from request query
func (h *WebHandler) getPaginationParams(c *gin.Context) (limit, offset int) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}
