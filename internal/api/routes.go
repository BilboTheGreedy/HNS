package api

import (
	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/internal/dns"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the API routes
func SetupRouter(
	router *gin.Engine,
	genService *service.GeneratorService,
	resService *service.ReservationService,
	seqService *service.SequenceService,
	userRepo repository.UserRepository,
	jwtManager *auth.JWTManager,
	apiKeyManager *auth.APIKeyManager,
	dnsChecker *dns.DNSChecker,
) {
	// Create handlers
	apiHandler := NewAPIHandler(genService, resService, seqService, dnsChecker)
	authHandler := NewAuthHandler(userRepo, jwtManager, apiKeyManager)

	// Public routes
	router.GET("/health", apiHandler.HealthCheck)

	// Auth routes
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.RegisterUser)
		authRoutes.POST("/login", authHandler.Login)
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
			templates.DELETE("/:id", AuthMiddleware(jwtManager, apiKeyManager, "admin"), apiHandler.DeleteTemplate)
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
		dnsRoutes := api.Group("/dns")
		{
			dnsRoutes.GET("/check/:hostname", apiHandler.CheckHostnameDNS)
			dnsRoutes.POST("/scan", apiHandler.ScanDNS)
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
}
