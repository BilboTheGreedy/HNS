package web

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/internal/dns"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/web/handlers"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/bilbothegreedy/HNS/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the web router
func SetupRouter(
	router *gin.Engine,
	userRepo repository.UserRepository,
	hostnameRepo repository.HostnameRepository,
	templateRepo repository.TemplateRepository,
	jwtManager *auth.JWTManager,
	dnsChecker *dns.DNSChecker,
) {
	// 1. Set up template renderer FIRST - before any routes are registered
	setupTemplates(router)

	// 2. Set up session store
	helpers.SetupSessionStore(router)

	// 3. Serve static files
	router.Static("/static", "./internal/web/static")

	// 4. Create middleware
	authMiddleware := middleware.NewAuthMiddleware(userRepo)
	router.Use(authMiddleware.LoadUser())

	// 5. Create handlers
	authHandler := handlers.NewAuthHandler(userRepo)
	dashboardHandler := handlers.NewDashboardHandler(hostnameRepo, templateRepo)
	baseHandler := handlers.NewBaseHandler()

	// 6. Register routes AFTER template setup
	// Unprotected routes
	router.GET("/login", authHandler.ShowLogin)
	router.POST("/login", authHandler.Login)
	router.GET("/logout", authHandler.Logout)

	// Redirect root to dashboard if logged in, otherwise to login
	router.GET("/", func(c *gin.Context) {
		if helpers.IsAuthenticated(c) {
			c.Redirect(302, "/dashboard")
		} else {
			c.Redirect(302, "/login")
		}
	})

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(authMiddleware.RequireAuth())
	{
		// Dashboard
		authorized.GET("/dashboard", dashboardHandler.Show)

		// Add more routes here as needed
	}

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(authMiddleware.RequireAuth(), authMiddleware.RequireAdmin())
	{
		// Add admin routes here
	}

	// Error handlers
	router.NoRoute(baseHandler.NotFound)
}

// setupTemplates configures the template renderer
func setupTemplates(router *gin.Engine) {
	// Set up template functions
	funcMap := template.FuncMap{
		"formatDate": helpers.FormatDate,
		// Add more template functions as needed
	}

	// Load templates with custom template functions
	templatesPattern := []string{
		"./internal/web/templates/layouts/*.html",
		"./internal/web/templates/partials/*.html",
		"./internal/web/templates/pages/*.html",
	}

	tmpl, err := loadTemplates(templatesPattern, funcMap)
	if err != nil {
		panic(err)
	}

	router.SetHTMLTemplate(tmpl)
}

// loadTemplates loads templates from the given patterns with functions
func loadTemplates(patterns []string, funcMap template.FuncMap) (*template.Template, error) {
	var allFiles []string

	// Get all template files
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, files...)
	}

	// Create template with functions
	tmpl := template.New("").Funcs(funcMap)

	// Parse all template files
	for _, file := range allFiles {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		name := filepath.Base(file)

		// If the file is a page template (not a partial or layout),
		// use both its content and base content for the template
		if strings.HasPrefix(name, "404") || strings.HasPrefix(name, "500") || strings.HasPrefix(name, "403") {
			// For error pages, use the special name for direct rendering
			t := tmpl.New(name)
			_, err = t.Parse(string(b))
			if err != nil {
				return nil, err
			}
		} else {
			_, err = tmpl.New(name).Parse(string(b))
			if err != nil {
				return nil, err
			}
		}
	}

	return tmpl, nil
}
