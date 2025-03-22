package web

import (
	"os"
	"path/filepath"

	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/internal/dns"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/service"
	"github.com/bilbothegreedy/HNS/internal/web/handlers"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/bilbothegreedy/HNS/internal/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// SetupWebRouter sets up the web router and handlers
func SetupWebRouter(
	router *gin.Engine,
	userRepo repository.UserRepository,
	templateRepo repository.TemplateRepository,
	hostnameRepo repository.HostnameRepository,
	jwtManager *auth.JWTManager,
	generatorSvc *service.GeneratorService,
	reservationSvc *service.ReservationService,
	dnsChecker *dns.DNSChecker,
	sequenceSvc *service.SequenceService,
) {
	// Setup session store
	helpers.SetupSessionStore(router)

	// Setup templates
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get working directory")
	}
	helpers.LoadTemplates(router, wd)

	// Setup static files
	staticPath := filepath.Join(wd, "internal", "web", "static")
	log.Info().Str("static_path", staticPath).Msg("Setting up static files")

	// Check if static path exists
	if _, err := os.Stat(staticPath); os.IsNotExist(err) {
		log.Error().Str("path", staticPath).Msg("Static files directory doesn't exist")
	}

	router.Static("/static", staticPath)

	// Create auth middleware
	authMw := middleware.NewAuthMiddleware(userRepo)

	// Create API key manager
	apiKeyManager := handlers.CreateAPIKeyManager(userRepo)

	// Create handlers
	baseHandler := handlers.NewBaseHandler()
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	dashboardHandler := handlers.NewDashboardHandler(hostnameRepo, templateRepo)
	hostnameHandler := handlers.NewHostnameHandler(hostnameRepo, templateRepo, generatorSvc, reservationSvc)
	templateHandler := handlers.NewTemplateHandler(templateRepo, generatorSvc)
	dnsHandler := handlers.NewDNSHandler(templateRepo, dnsChecker, generatorSvc)
	userHandler := handlers.NewUserHandler(userRepo, hostnameRepo)
	apiKeyHandler := handlers.NewAPIKeyHandler(userRepo, apiKeyManager)

	// Setup public routes
	router.GET("/", dashboardHandler.Home)
	router.GET("/login", authHandler.LoginPage)
	router.POST("/login", authHandler.Login)
	router.GET("/register", authHandler.RegisterPage)
	router.POST("/register", authHandler.Register)
	router.GET("/logout", authHandler.Logout)

	// Error pages
	router.GET("/404", baseHandler.NotFound)
	router.GET("/403", baseHandler.Forbidden)

	// Add a test route for debugging
	router.GET("/test", func(c *gin.Context) {
		c.HTML(200, "pages/login.html", gin.H{
			"Title":       "Test Page",
			"CurrentYear": helpers.GetCurrentYear(),
		})
	})
	router.Use(func(c *gin.Context) {
		// Extract session data and add to context
		helpers.AddContextUserData(c)
		c.Next()
	})
	// Protected routes
	auth := router.Group("/")
	auth.Use(authMw.AuthRequired())
	{
		// Dashboard
		auth.GET("/dashboard", dashboardHandler.Dashboard)

		// Hostname routes
		auth.GET("/hostnames", hostnameHandler.HostnameList)
		auth.GET("/hostnames/generate", hostnameHandler.GenerateHostnamePage)
		auth.POST("/hostnames/generate", hostnameHandler.GenerateHostname)
		auth.GET("/hostnames/reserve", hostnameHandler.ReserveHostnamePage)
		auth.POST("/hostnames/reserve", hostnameHandler.ReserveHostname)
		auth.GET("/hostnames/:id", hostnameHandler.HostnameDetail)
		auth.GET("/hostnames/:id/commit", hostnameHandler.CommitHostname)
		auth.GET("/hostnames/:id/release", hostnameHandler.ReleaseHostname)

		// Template routes
		auth.GET("/templates", templateHandler.TemplateList)
		auth.GET("/templates/:id", templateHandler.TemplateDetail)

		// DNS routes
		auth.GET("/dns", dnsHandler.DNSTools)
		auth.GET("/dns/check", dnsHandler.DNSCheck)
		auth.GET("/dns/check/:hostname", dnsHandler.DNSCheckHostname)
		auth.GET("/dns/scan", dnsHandler.DNSScan)
		auth.POST("/dns/scan", dnsHandler.DNSScanSubmit)
		auth.POST("/dns/discover", dnsHandler.DNSDiscover)

		// User profile
		auth.GET("/profile", userHandler.UserProfile)
		auth.POST("/profile/update", userHandler.UpdateProfile)
		auth.POST("/profile/change-password", userHandler.ChangePassword)

		// API keys
		auth.GET("/api-keys", apiKeyHandler.APIKeysList)
		auth.POST("/api-keys", apiKeyHandler.CreateAPIKey)
		auth.GET("/api-keys/:id/delete", apiKeyHandler.DeleteAPIKey)

		// Admin routes
		admin := auth.Group("/admin")
		admin.Use(authMw.AdminRequired())
		{
			// User management
			admin.GET("/users", userHandler.UsersList)
			admin.POST("/users", userHandler.CreateUser)
			admin.POST("/users/update", userHandler.UpdateUser)
			admin.GET("/users/:id", userHandler.ViewUser)
			admin.GET("/users/:id/delete", userHandler.DeleteUser)

			// Template management
			admin.GET("/templates/new", templateHandler.NewTemplate)
			admin.POST("/templates", templateHandler.CreateTemplate)
			admin.GET("/templates/:id/edit", templateHandler.EditTemplate)
			admin.POST("/templates/:id", templateHandler.UpdateTemplate)
			admin.GET("/templates/:id/delete", templateHandler.DeleteTemplate)
		}
	}

	// 404 handler
	router.NoRoute(baseHandler.NotFound)
}
