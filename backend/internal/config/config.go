package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	Port        string `env:"PORT" envDefault:"8080"`
	
	// Database - PostgreSQL
	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	PostgresPort     string `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"directus"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"directus123"`
	PostgresDB       string `env:"POSTGRES_DB" envDefault:"directus"`
	PostgresSSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`
	
	// Redis
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`
	
	// JWT
	JWTSecret              string `env:"JWT_SECRET" envDefault:"your-secret-key"`
	JWTAccessExpiration    int    `env:"JWT_ACCESS_EXPIRATION" envDefault:"15"`    // minutes
	JWTRefreshExpiration   int    `env:"JWT_REFRESH_EXPIRATION" envDefault:"168"`  // hours (7 days)
	
	// CORS
	CORSOrigins string `env:"CORS_ORIGINS" envDefault:"http://localhost:5173,http://localhost:3000"`
	
	// File Upload
	MaxFileSize   int64  `env:"MAX_FILE_SIZE" envDefault:"10485760"` // 10MB
	UploadPath    string `env:"UPLOAD_PATH" envDefault:"./uploads"`
	
	// i18n
	DefaultLanguage string `env:"DEFAULT_LANGUAGE" envDefault:"en"`
	
	// Rate Limiting
	RateLimitRequests int `env:"RATE_LIMIT_REQUESTS" envDefault:"100"`
	RateLimitWindow   int `env:"RATE_LIMIT_WINDOW" envDefault:"60"` // seconds
	
	// Email Configuration
	SMTPHost           string `env:"SMTP_HOST" envDefault:"smtp.gmail.com"`
	SMTPPort           int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUsername       string `env:"SMTP_USERNAME" envDefault:""`
	SMTPPassword       string `env:"SMTP_PASSWORD" envDefault:""`
	FromEmail          string `env:"FROM_EMAIL" envDefault:"noreply@eralove.com"`
	FromName           string `env:"FROM_NAME" envDefault:"EraLove"`
	EnableEmailVerify  bool   `env:"ENABLE_EMAIL_VERIFY" envDefault:"false"`
	
	// Frontend URL for email links
	FrontendURL string `env:"FRONTEND_URL" envDefault:"http://localhost:3000"`
	
	// Storage Configuration
	StorageProvider     string `env:"STORAGE_PROVIDER" envDefault:"local"`        // local, s3
	StorageRegion       string `env:"STORAGE_REGION" envDefault:"us-east-1"`
	StorageBucket       string `env:"STORAGE_BUCKET" envDefault:"eralove-uploads"`
	StorageAccessKeyID  string `env:"STORAGE_ACCESS_KEY_ID" envDefault:""`
	StorageSecretKey    string `env:"STORAGE_SECRET_KEY" envDefault:""`
	StorageEndpoint     string `env:"STORAGE_ENDPOINT" envDefault:""`             // For MinIO or custom S3
	StorageUseSSL       bool   `env:"STORAGE_USE_SSL" envDefault:"true"`
	StorageBaseURL      string `env:"STORAGE_BASE_URL" envDefault:"http://localhost:8080"` // For public file access
	
	// Directus CMS Configuration
	DirectusURL         string `env:"DIRECTUS_URL" envDefault:"http://localhost:8055"`
	DirectusAdminEmail  string `env:"DIRECTUS_ADMIN_EMAIL" envDefault:"admin@eralove.com"`
	DirectusAdminPass   string `env:"DIRECTUS_ADMIN_PASSWORD" envDefault:"Admin@123456"`
	DirectusStaticToken string `env:"DIRECTUS_STATIC_TOKEN" envDefault:""`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist
		fmt.Println("No .env file found, using environment variables")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// validate checks if all required configuration is present
func (c *Config) validate() error {
	if c.JWTSecret == "your-secret-key" && c.Environment == "production" {
		return fmt.Errorf("JWT_SECRET must be set in production")
	}

	// PostgreSQL is now the primary database
	if c.PostgresHost == "" {
		return fmt.Errorf("POSTGRES_HOST is required")
	}

	return nil
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// GetPort returns the port with colon prefix
func (c *Config) GetPort() string {
	return ":" + c.Port
}

// GetRedisDB returns Redis DB as integer
func (c *Config) GetRedisDB() int {
	if db, err := strconv.Atoi(os.Getenv("REDIS_DB")); err == nil {
		return db
	}
	return c.RedisDB
}
