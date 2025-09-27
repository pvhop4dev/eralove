package storage

import (
	"fmt"
	"strings"

	"github.com/eralove/eralove-backend/internal/domain"
	"go.uber.org/zap"
)

// StorageProvider represents different storage providers
type StorageProvider string

const (
	ProviderLocal StorageProvider = "local"
	ProviderMinIO StorageProvider = "minio"
	ProviderS3    StorageProvider = "s3"
)

// Factory creates storage services based on configuration
type Factory struct {
	logger *zap.Logger
}

// NewFactory creates a new storage factory
func NewFactory(logger *zap.Logger) *Factory {
	return &Factory{
		logger: logger,
	}
}

// CreateStorage creates a storage service based on the provider configuration
func (f *Factory) CreateStorage(config *domain.StorageConfig) (domain.StorageService, error) {
	provider := StorageProvider(strings.ToLower(config.Provider))

	f.logger.Info("Creating storage service",
		zap.String("provider", string(provider)),
		zap.String("bucket", config.Bucket),
		zap.String("region", config.Region))

	switch provider {
	case ProviderLocal:
		return f.createLocalStorage(config)
	case ProviderMinIO, ProviderS3:
		return f.createMinIOStorage(config)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", provider)
	}
}

// createLocalStorage creates a local filesystem storage service
func (f *Factory) createLocalStorage(config *domain.StorageConfig) (domain.StorageService, error) {
	// For local storage, use bucket as the base path
	basePath := config.Bucket
	if basePath == "" {
		basePath = "./uploads"
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	f.logger.Info("Creating local storage",
		zap.String("base_path", basePath),
		zap.String("base_url", baseURL))

	return NewLocalStorage(basePath, baseURL, f.logger)
}

// createMinIOStorage creates a MinIO/S3 compatible storage service
func (f *Factory) createMinIOStorage(config *domain.StorageConfig) (domain.StorageService, error) {
	// Validate required fields for MinIO/S3
	if config.AccessKeyID == "" {
		return nil, fmt.Errorf("access key ID is required for MinIO/S3 storage")
	}
	if config.SecretAccessKey == "" {
		return nil, fmt.Errorf("secret access key is required for MinIO/S3 storage")
	}
	if config.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required for MinIO/S3 storage")
	}

	f.logger.Info("Creating MinIO/S3 storage",
		zap.String("endpoint", config.Endpoint),
		zap.String("bucket", config.Bucket),
		zap.String("region", config.Region),
		zap.Bool("use_ssl", config.UseSSL))

	return NewMinIOStorage(config, f.logger)
}

// GetStorageConfig creates storage config from individual parameters
func GetStorageConfig(provider, region, bucket, accessKey, secretKey, endpoint, baseURL string, useSSL bool) *domain.StorageConfig {
	return &domain.StorageConfig{
		Provider:        provider,
		Region:          region,
		Bucket:          bucket,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Endpoint:        endpoint,
		UseSSL:          useSSL,
		BaseURL:         baseURL,
	}
}

// ValidateConfig validates storage configuration
func ValidateConfig(config *domain.StorageConfig) error {
	if config.Provider == "" {
		return fmt.Errorf("storage provider is required")
	}

	provider := StorageProvider(strings.ToLower(config.Provider))

	switch provider {
	case ProviderLocal:
		// Local storage doesn't need much validation
		return nil
	case ProviderMinIO, ProviderS3:
		if config.AccessKeyID == "" {
			return fmt.Errorf("access key ID is required for %s storage", provider)
		}
		if config.SecretAccessKey == "" {
			return fmt.Errorf("secret access key is required for %s storage", provider)
		}
		if config.Bucket == "" {
			return fmt.Errorf("bucket name is required for %s storage", provider)
		}
		if config.Region == "" && config.Endpoint == "" {
			return fmt.Errorf("either region or endpoint is required for %s storage", provider)
		}
		return nil
	default:
		return fmt.Errorf("unsupported storage provider: %s", provider)
	}
}

// GetDefaultConfig returns default storage configuration for development
func GetDefaultConfig() *domain.StorageConfig {
	return &domain.StorageConfig{
		Provider: "local",
		Bucket:   "./uploads",
		BaseURL:  "http://localhost:8080",
		UseSSL:   false,
	}
}

// GetMinIOConfig returns MinIO configuration for development
func GetMinIOConfig() *domain.StorageConfig {
	return &domain.StorageConfig{
		Provider:        "minio",
		Region:          "us-east-1",
		Bucket:          "eralove-uploads",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		Endpoint:        "localhost:9000",
		UseSSL:          false,
		BaseURL:         "http://localhost:9000",
	}
}

// GetS3Config returns AWS S3 configuration template
func GetS3Config(region, bucket, accessKey, secretKey string) *domain.StorageConfig {
	return &domain.StorageConfig{
		Provider:        "s3",
		Region:          region,
		Bucket:          bucket,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		UseSSL:          true,
		BaseURL:         fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region),
	}
}
