package api

import (
	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/internal/config"
	"github.com/bilbothegreedy/HNS/internal/dns"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the API routes
func SetupRouter(
	cfg *config.Config,
	generatorService *service.GeneratorService,
	reservationService *service.ReservationService,
	sequenceService *service.SequenceService,
	userRepo repository.UserRepository,
) *gin.Engine {
	// Create JWT manager
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiration)

	// Create API key manager
	apiKeyManager := auth.NewAPIKeyManager(userRepo, cfg.Auth.APIKeyExpiration)

	// Create DNS checker
	dnsChecker := dns.NewDNSChecker(cfg.DNS)

	// Create handlers
	apiHandler := NewAPIHandler(generatorService, reservationService, sequenceService, dnsChecker)
	authHandler := NewAuthHandler(userRepo, jwtManager, apiKeyManager)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(LoggerMiddleware())
	router.Use(RecoveryMiddleware())
	router.Use(CORSMiddleware())

	// Public routes
	router.GET("/health", apiHandler.HealthCheck)

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.RegisterUser)
		auth.POST("/login", authHandler.Login)
	}

	// API routes requiring authentication
	api := router.Group("/api")
	api.Use(AuthMiddleware(jwtManager, apiKeyManager, "read"))
	{
		// Template routes
		templates := api.Group("/templates")
		{
			templates.GET("", apiHandler.GetTemplates)
			templates.GET("/:id", apiHandler.GetTemplate)
			templates.POST("", AuthMiddleware(jwtManager, apiKeyManager, "admin"), apiHandler.CreateTemplate)
		}

		// Hostname routes
		hostnames := api.Group("/hostnames")
		{
			hostnames.POST("/generate", apiHandler.GenerateHostname)
			hostnames.POST("/reserve", AuthMiddleware(jwtManager, apiKeyManager, "reserve"), apiHandler.ReserveHostname)
			hostnames.POST("/commit", AuthMiddleware(jwtManager, apiKeyManager, "commit"), apiHandler.CommitHostname)
			hostnames.POST("/release", AuthMiddleware(jwtManager, apiKeyManager, "release"), apiHandler.ReleaseHostname)
			hostnames.GET("/reserved", apiHandler.GetReservedHostnames)
			hostnames.GET("/committed", apiHandler.GetCommittedHostnames)
			hostnames.GET("/:id", apiHandler.GetHostname)
			hostnames.GET("", apiHandler.SearchHostnames)
		}

		// Sequence routes
		sequences := api.Group("/sequences")
		{
			sequences.GET("/next/:templateID", apiHandler.GetNextSequenceNumber)
		}

		// DNS routes
		dns := api.Group("/dns")
		{
			dns.GET("/check/:hostname", apiHandler.CheckHostnameDNS)
			dns.POST("/scan", apiHandler.ScanDNS)
		}

		// User routes
		users := api.Group("/users")
		users.Use(RoleMiddleware("admin"))
		{
			users.GET("", authHandler.GetUsers)
			users.GET("/:id", authHandler.GetUser)
			users.PUT("/:id", authHandler.UpdateUser)
			users.DELETE("/:id", authHandler.DeleteUser)
		}

		// API key routes
		apiKeys := api.Group("/apikeys")
		{
			apiKeys.GET("", authHandler.GetApiKeys)
			apiKeys.POST("", authHandler.CreateApiKey)
			apiKeys.DELETE("/:id", authHandler.DeleteApiKey)
		}
	}

	return router
}
