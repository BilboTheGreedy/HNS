package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bilbothegreedy/HNS/internal/api"
	"github.com/bilbothegreedy/HNS/internal/auth"
	"github.com/bilbothegreedy/HNS/internal/config"
	"github.com/bilbothegreedy/HNS/internal/db/migration"
	"github.com/bilbothegreedy/HNS/internal/dns"
	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/repository/postgres"
	"github.com/bilbothegreedy/HNS/internal/service"
	"github.com/bilbothegreedy/HNS/internal/web"
	"github.com/bilbothegreedy/HNS/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func ensureAdminUserExists(userRepo repository.UserRepository) {
	ctx := context.Background()

	// Try to get the admin user
	user, err := userRepo.GetByUsername(ctx, "admin")
	if err == nil && user != nil {
		log.Info().Msg("Admin user already exists")
		return
	}

	// Create admin user
	log.Info().Msg("Creating default admin user")

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to hash admin password")
		return
	}

	// Create the user
	adminUser := &models.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: string(hashedPassword),
		FirstName:    "Admin",
		LastName:     "User",
		Role:         models.RoleAdmin,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := userRepo.Create(ctx, adminUser); err != nil {
		log.Error().Err(err).Msg("Failed to create admin user - may already exist")
		return
	}

	log.Info().Msg("Default admin user created successfully")
}

func main() {
	// Initialize logger
	utils.InitLogger(zerolog.InfoLevel)
	log.Info().Msg("Starting Hostname Naming System (HNS)")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database connection
	db, err := postgres.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Run migrations if enabled
	if cfg.Database.RunMigrations {
		log.Info().Msg("Running database migrations")
		migrator := migration.NewMigration(cfg.Database, "migrations")
		if err := migrator.Migrate(); err != nil {
			log.Fatal().Err(err).Msg("Database migration failed")
		}
		log.Info().Msg("Database migrations completed")
	}

	// Create repositories
	hostRepo := postgres.NewHostnameRepository(db)
	templateRepo := postgres.NewTemplateRepository(db)
	userRepo := postgres.NewUserRepository(db)

	// Ensure admin user exists
	ensureAdminUserExists(userRepo)

	// Create services
	genService := service.NewGeneratorService(templateRepo)
	resService := service.NewReservationService(hostRepo, templateRepo)
	seqService := service.NewSequenceService(hostRepo)

	// Create auth components
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiration)
	apiKeyManager := auth.NewAPIKeyManager(userRepo, cfg.Auth.APIKeyExpiration)

	// Create DNS checker
	dnsChecker := dns.NewDNSChecker(cfg.DNS)
	// Add this to your main function or a debug endpoint
	templatePaths := []string{
		"./internal/web/templates/layouts/base/base.html",
		"./internal/web/templates/layouts/pages/login.html",
		"./internal/web/templates/layouts/partials/navbar.html",
	}

	for _, path := range templatePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Error().Str("path", path).Msg("Template file doesn't exist")
		} else {
			log.Info().Str("path", path).Msg("Template file exists")
		}
	}
	// Set up Gin with appropriate mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup API routes under /api path

	api.SetupRouter(
		router, // Pass the entire engine
		genService,
		resService,
		seqService,
		userRepo,
		jwtManager,
		apiKeyManager,
		dnsChecker,
	)

	// Setup Web routes for the UI
	web.SetupRouter(
		router,
		userRepo,
		hostRepo,
		templateRepo,
		jwtManager,
		dnsChecker,
	)

	// Start server in a goroutine
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	go func() {
		log.Info().Msgf("Starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited properly")
}
