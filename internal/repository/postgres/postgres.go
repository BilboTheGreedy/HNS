package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/bilbothegreedy/HNS/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// DB is a wrapper around pgxpool.Pool
type DB struct {
	pool *pgxpool.Pool
}

// Pool returns the underlying connection pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg config.DatabaseConfig) (*DB, error) {
	// Build connection string
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Set pool size and timeouts
	poolConfig.MaxConns = int32(cfg.PoolSize)
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg("Connected to PostgreSQL database")
	return &DB{pool: pool}, nil
}

// ExecTx executes a function within a transaction
func (db *DB) ExecTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// QueryRow is a wrapper around pgxpool.Pool.QueryRow
func (db *DB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, sql, args...)
}

// Query is a wrapper around pgxpool.Pool.Query
func (db *DB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, args...)
}

// Exec is a wrapper around pgxpool.Pool.Exec
func (db *DB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, sql, args...)
}

