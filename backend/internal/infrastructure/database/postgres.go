package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// PostgresDB wraps the PostgreSQL database connection
type PostgresDB struct {
	DB *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(config PostgresConfig) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{DB: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

// Health checks the database health
func (p *PostgresDB) Health() error {
	return p.DB.Ping()
}

// GetDB returns the underlying sql.DB instance
func (p *PostgresDB) GetDB() *sql.DB {
	return p.DB
}
