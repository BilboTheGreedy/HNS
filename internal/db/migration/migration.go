package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bilbothegreedy/HNS/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // Import the lib/pq driver
	"github.com/rs/zerolog/log"
)

// Migration handles database migrations
type Migration struct {
	migrationsPath string
	dbConfig       config.DatabaseConfig
}

// NewMigration creates a new Migration
func NewMigration(dbConfig config.DatabaseConfig, migrationsPath string) *Migration {
	return &Migration{
		migrationsPath: migrationsPath,
		dbConfig:       dbConfig,
	}
}

// Migrate runs all pending migrations
func (m *Migration) Migrate() error {
	log.Info().Str("path", m.migrationsPath).Msg("Running database migrations")

	// Verify the migrations path exists
	if _, err := os.Stat(m.migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", m.migrationsPath)
	}

	// Try to create database if it doesn't exist
	if err := m.ensureDatabaseExists(); err != nil {
		log.Warn().Err(err).Msg("Failed to ensure database exists")
	}

	// Convert migrationsPath to absolute path
	absPath, err := filepath.Abs(m.migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Build DSN for database connection
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		m.dbConfig.User, m.dbConfig.Password, m.dbConfig.Host,
		m.dbConfig.Port, m.dbConfig.DBName, m.dbConfig.SSLMode)

	// Create migration instance
	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(absPath))
	instance, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer instance.Close()

	// Log the connection and source URLs (not including password)
	log.Debug().
		Str("source", sourceURL).
		Str("database", fmt.Sprintf("postgres://%s:****@%s:%d/%s?sslmode=%s",
			m.dbConfig.User, m.dbConfig.Host, m.dbConfig.Port, m.dbConfig.DBName, m.dbConfig.SSLMode)).
		Msg("Migration configuration")

	// Run migrations
	if err := instance.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("No migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Info().Msg("Migrations applied successfully")
	return nil
}

// CreateMigration creates a new migration file
func (m *Migration) CreateMigration(name string) error {
	// Format name (convert spaces to underscores, lowercase)
	name = strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	// Get current timestamp as version
	version := getTimestampVersion()

	// Create migration files
	upFilename := fmt.Sprintf("%s/%s_%s.up.sql", m.migrationsPath, version, name)
	downFilename := fmt.Sprintf("%s/%s_%s.down.sql", m.migrationsPath, version, name)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(m.migrationsPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Create up migration file
	upFile, err := os.Create(upFilename)
	if err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}
	defer upFile.Close()

	// Write a comment to the up file
	upFile.WriteString(fmt.Sprintf("-- Migration: %s\n\n", name))

	// Create down migration file
	downFile, err := os.Create(downFilename)
	if err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}
	defer downFile.Close()

	// Write a comment to the down file
	downFile.WriteString(fmt.Sprintf("-- Revert: %s\n\n", name))

	log.Info().
		Str("up", upFilename).
		Str("down", downFilename).
		Msg("Migration files created")

	return nil
}

// getTimestampVersion returns a timestamp for versioning
func getTimestampVersion() string {
	// Use Unix timestamp in seconds
	return fmt.Sprintf("%d", time.Now().Unix())
}

// ensureDatabaseExists attempts to create the database if it doesn't exist
func (m *Migration) ensureDatabaseExists() error {
	// Build DSN for postgres database (not the target database)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=%s",
		m.dbConfig.User, m.dbConfig.Password, m.dbConfig.Host,
		m.dbConfig.Port, m.dbConfig.SSLMode)

	// Open connection to postgres database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRow(query, m.dbConfig.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// If database doesn't exist, create it
	if !exists {
		log.Info().Str("dbname", m.dbConfig.DBName).Msg("Database does not exist, creating...")
		_, err = db.Exec("CREATE DATABASE " + m.dbConfig.DBName)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Info().Str("dbname", m.dbConfig.DBName).Msg("Database created successfully")
	} else {
		log.Info().Str("dbname", m.dbConfig.DBName).Msg("Database already exists")
	}

	return nil
}
