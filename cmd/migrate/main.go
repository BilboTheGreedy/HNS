package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bilbothegreedy/HNS/internal/config"
	"github.com/bilbothegreedy/HNS/internal/db/migration"
	"github.com/bilbothegreedy/HNS/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	utils.InitLogger(zerolog.InfoLevel)

	// Command line flags
	migrationsPath := flag.String("path", "migrations", "Path to migration files")
	createFlag := flag.String("create", "", "Create a new migration (provide migration name)")
	migrateFlag := flag.Bool("up", false, "Run all pending migrations")

	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Print database configuration
	log.Info().
		Str("host", cfg.Database.Host).
		Int("port", cfg.Database.Port).
		Str("user", cfg.Database.User).
		Str("dbname", cfg.Database.DBName).
		Str("sslmode", cfg.Database.SSLMode).
		Msg("Database configuration")
		
	// Create migrations directory if it doesn't exist
	if _, err := os.Stat(*migrationsPath); os.IsNotExist(err) {
		log.Info().Msgf("Creating migrations directory: %s", *migrationsPath)
		if err := os.MkdirAll(*migrationsPath, 0755); err != nil {
			log.Fatal().Err(err).Msgf("Failed to create migrations directory: %s", *migrationsPath)
		}
	}

	// Create migration
	if *createFlag != "" {
		// Initialize migration without database connection
		m := migration.NewMigration(cfg.Database, *migrationsPath)

		// Create new migration files
		if err := m.CreateMigration(*createFlag); err != nil {
			log.Fatal().Err(err).Msg("Failed to create migration")
		}

		fmt.Printf("Migration created: %s\n", *createFlag)
		return
	}

	// Run migrations
	if *migrateFlag {
		// Initialize migration
		m := migration.NewMigration(cfg.Database, *migrationsPath)

		// Run migrations
		if err := m.Migrate(); err != nil {
			log.Fatal().Err(err).Msg("Failed to run migrations")
		}

		fmt.Println("Migrations completed successfully")
		return
	}

	// No command provided
	fmt.Println("No command provided. Use -create or -up.")
	os.Exit(1)
}
