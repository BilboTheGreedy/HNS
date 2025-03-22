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
	"github.com/bilbothegreedy/HNS/internal/repository/postgres"
	"github.com/bilbothegreedy/HNS/internal/service"
	"github.com/bilbothegreedy/HNS/internal/web"
	"github.com/bilbothegreedy/HNS/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	utils.InitLogger(zerolog.InfoLevel)
	log.Info().Msg("Starting server name generator application")

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

	// Create services
	genService := service.NewGeneratorService(templateRepo)
	resService := service.NewReservationService(hostRepo, templateRepo)
	seqService := service.NewSequenceService(hostRepo)

	// Create auth components
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiration)
	apiKeyManager := auth.NewAPIKeyManager(userRepo, cfg.Auth.APIKeyExpiration)

	// Create DNS checker
	dnsChecker := dns.NewDNSChecker(cfg.DNS)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode) // Use ReleaseMode for production
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup API routes
	api.SetupRouter(
		router,
		genService,
		resService,
		seqService,
		userRepo,
		jwtManager,
		apiKeyManager,
		dnsChecker,
	)

	// Setup web routes (using our refactored web package)
	web.SetupWebRouter(
		router,
		userRepo,
		templateRepo,
		hostRepo,
		jwtManager,
		genService,
		resService,
		dnsChecker,
		seqService,
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited properly")
}
