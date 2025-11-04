package directus

import (
	"fmt"
	"os"
)

// Config holds Directus configuration
type Config struct {
	URL         string
	AdminEmail  string
	AdminPass   string
	StaticToken string
}

// LoadConfig loads Directus configuration from environment variables
func LoadConfig() (*Config, error) {
	url := os.Getenv("DIRECTUS_URL")
	if url == "" {
		return nil, fmt.Errorf("DIRECTUS_URL is required")
	}

	return &Config{
		URL:         url,
		AdminEmail:  os.Getenv("DIRECTUS_ADMIN_EMAIL"),
		AdminPass:   os.Getenv("DIRECTUS_ADMIN_PASSWORD"),
		StaticToken: os.Getenv("DIRECTUS_STATIC_TOKEN"),
	}, nil
}

// NewClientFromConfig creates a new Directus client from config
func NewClientFromConfig(cfg *Config) (*Client, error) {
	client := NewClient(cfg.URL)

	// Use static token if available
	if cfg.StaticToken != "" {
		client.SetToken(cfg.StaticToken)
		return client, nil
	}

	// Otherwise authenticate with email/password
	if cfg.AdminEmail != "" && cfg.AdminPass != "" {
		if err := client.Authenticate(cfg.AdminEmail, cfg.AdminPass); err != nil {
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
		return client, nil
	}

	// Return unauthenticated client (for public endpoints)
	return client, nil
}
