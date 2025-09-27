package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/eralove/eralove-backend/internal/domain"
	"go.uber.org/zap"
)

// MinIOStorage implements StorageService using MinIO (compatible with S3)
type MinIOStorage struct {
	client *minio.Client
	config *domain.StorageConfig
	logger *zap.Logger
}

// NewMinIOStorage creates a new MinIO storage service (compatible with S3 and MinIO)
func NewMinIOStorage(config *domain.StorageConfig, logger *zap.Logger) (domain.StorageService, error) {
	// Determine endpoint
	endpoint := config.Endpoint
	if endpoint == "" {
		// Default to AWS S3 endpoint
		endpoint = fmt.Sprintf("s3.%s.amazonaws.com", config.Region)
	}

	// Remove protocol from endpoint if present
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	// Create MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, config.Bucket)
	if err != nil {
		logger.Warn("Failed to check bucket existence", zap.Error(err))
	} else if !exists {
		logger.Info("Creating bucket", zap.String("bucket", config.Bucket))
		err = client.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{Region: config.Region})
		if err != nil {
			logger.Warn("Failed to create bucket", zap.Error(err))
		}
	}

	return &MinIOStorage{
		client: client,
		config: config,
		logger: logger,
	}, nil
}

// Upload uploads a file to MinIO/S3
func (m *MinIOStorage) Upload(ctx context.Context, req *domain.UploadRequest) (*domain.FileInfo, error) {
	// Generate unique key
	key := m.generateKey(req.Folder, req.UserID, req.Filename)

	m.logger.Info("Starting MinIO upload",
		zap.String("key", key),
		zap.String("filename", req.Filename),
		zap.String("content_type", req.ContentType),
		zap.Int64("size", req.Size))

	// Upload to MinIO/S3
	info, err := m.client.PutObject(ctx, m.config.Bucket, key, req.File, req.Size, minio.PutObjectOptions{
		ContentType: req.ContentType,
	})
	if err != nil {
		m.logger.Error("Failed to upload to MinIO", zap.Error(err), zap.String("key", key))
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate public URL
	url := m.generatePublicURL(key)

	fileInfo := &domain.FileInfo{
		Key:         key,
		URL:         url,
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Size:        info.Size,
		UploadedAt:  time.Now(),
		Bucket:      m.config.Bucket,
	}

	m.logger.Info("MinIO upload successful",
		zap.String("key", key),
		zap.String("url", url),
		zap.Int64("size", info.Size))

	return fileInfo, nil
}

// Download generates a presigned URL for downloading
func (m *MinIOStorage) Download(ctx context.Context, req *domain.DownloadRequest) (string, error) {
	return m.GeneratePresignedDownloadURL(ctx, req.Key, req.Expiry)
}

// Delete removes a file from MinIO/S3
func (m *MinIOStorage) Delete(ctx context.Context, key string) error {
	m.logger.Info("Deleting file from MinIO", zap.String("key", key))

	err := m.client.RemoveObject(ctx, m.config.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		m.logger.Error("Failed to delete from MinIO", zap.Error(err), zap.String("key", key))
		return fmt.Errorf("failed to delete file: %w", err)
	}

	m.logger.Info("File deleted from MinIO successfully", zap.String("key", key))
	return nil
}

// GetFileInfo retrieves file information from MinIO/S3
func (m *MinIOStorage) GetFileInfo(ctx context.Context, key string) (*domain.FileInfo, error) {
	objInfo, err := m.client.StatObject(ctx, m.config.Bucket, key, minio.StatObjectOptions{})
	if err != nil {
		m.logger.Error("Failed to get file info from MinIO", zap.Error(err), zap.String("key", key))
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Extract filename from key
	filename := filepath.Base(key)

	fileInfo := &domain.FileInfo{
		Key:         key,
		URL:         m.generatePublicURL(key),
		Filename:    filename,
		ContentType: objInfo.ContentType,
		Size:        objInfo.Size,
		UploadedAt:  objInfo.LastModified,
		Bucket:      m.config.Bucket,
	}

	return fileInfo, nil
}

// ListFiles lists files in a folder
func (m *MinIOStorage) ListFiles(ctx context.Context, folder string, limit int) ([]*domain.FileInfo, error) {
	objectCh := m.client.ListObjects(ctx, m.config.Bucket, minio.ListObjectsOptions{
		Prefix:    folder,
		Recursive: true,
	})

	var files []*domain.FileInfo
	count := 0

	for object := range objectCh {
		if object.Err != nil {
			m.logger.Error("Failed to list files from MinIO", zap.Error(object.Err), zap.String("folder", folder))
			return nil, fmt.Errorf("failed to list files: %w", object.Err)
		}

		if count >= limit {
			break
		}

		filename := filepath.Base(object.Key)

		fileInfo := &domain.FileInfo{
			Key:        object.Key,
			URL:        m.generatePublicURL(object.Key),
			Filename:   filename,
			Size:       object.Size,
			UploadedAt: object.LastModified,
			Bucket:     m.config.Bucket,
		}
		files = append(files, fileInfo)
		count++
	}

	return files, nil
}

// GeneratePresignedUploadURL generates a presigned URL for direct upload
func (m *MinIOStorage) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedPutObject(ctx, m.config.Bucket, key, expiry)
	if err != nil {
		m.logger.Error("Failed to generate presigned upload URL", zap.Error(err), zap.String("key", key))
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	m.logger.Info("Generated presigned upload URL", zap.String("key", key), zap.Duration("expiry", expiry))
	return url.String(), nil
}

// GeneratePresignedDownloadURL generates a presigned URL for direct download
func (m *MinIOStorage) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.config.Bucket, key, expiry, nil)
	if err != nil {
		m.logger.Error("Failed to generate presigned download URL", zap.Error(err), zap.String("key", key))
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	m.logger.Info("Generated presigned download URL", zap.String("key", key), zap.Duration("expiry", expiry))
	return url.String(), nil
}

// generateKey creates a unique key for the file
func (m *MinIOStorage) generateKey(folder, userID, filename string) string {
	// Clean filename
	cleanFilename := strings.ReplaceAll(filename, " ", "_")

	// Add timestamp to make it unique
	timestamp := time.Now().Format("20060102_150405")

	// Extract file extension
	ext := filepath.Ext(cleanFilename)
	nameWithoutExt := strings.TrimSuffix(cleanFilename, ext)

	// Create unique filename
	uniqueFilename := fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)

	// Construct full key
	if folder == "" {
		folder = "uploads"
	}

	return fmt.Sprintf("%s/%s/%s", folder, userID, uniqueFilename)
}

// generatePublicURL creates a public URL for the file
func (m *MinIOStorage) generatePublicURL(key string) string {
	if m.config.BaseURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(m.config.BaseURL, "/"), key)
	}

	// Default MinIO/S3 URL format
	protocol := "https"
	if !m.config.UseSSL {
		protocol = "http"
	}

	endpoint := m.config.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("s3.%s.amazonaws.com", m.config.Region)
	}

	return fmt.Sprintf("%s://%s/%s/%s", protocol, endpoint, m.config.Bucket, key)
}
