package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"blog-platform/internal/infrastructure/config"
)

// Database wraps sqlx.DB with additional functionality
type Database struct {
	*sqlx.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config) (*Database, error) {
	db, err := sqlx.Connect("mysql", cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}

// Migrate runs database migrations
func (d *Database) Migrate(migrationsPath string) error {
	// This will be implemented when we integrate with golang-migrate
	// For now, we'll return nil as migrations will be run separately
	return nil
}
