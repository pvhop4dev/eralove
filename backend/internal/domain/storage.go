package domain

import (
	"context"
	"errors"
	"io"
	"time"
)

// FileInfo represents information about an uploaded file
type FileInfo struct {
	Key         string    `json:"key"`          // S3 object key
	URL         string    `json:"url"`          // Public URL to access the file
	Filename    string    `json:"filename"`     // Original filename
	ContentType string    `json:"content_type"` // MIME type
	Size        int64     `json:"size"`         // File size in bytes
	UploadedAt  time.Time `json:"uploaded_at"`  // Upload timestamp
	Bucket      string    `json:"bucket"`       // S3 bucket name
}

// UploadRequest represents a file upload request
type UploadRequest struct {
	File        io.Reader `json:"-"`            // File content
	Filename    string    `json:"filename"`     // Original filename
	ContentType string    `json:"content_type"` // MIME type
	Size        int64     `json:"size"`         // File size
	Folder      string    `json:"folder"`       // Folder/prefix in S3
	UserID      string    `json:"user_id"`      // User ID for organization
}

// DownloadRequest represents a file download request
type DownloadRequest struct {
	Key    string        `json:"key"`     // S3 object key
	Expiry time.Duration `json:"expiry"`  // URL expiry duration
}

// StorageService defines the interface for file storage operations
type StorageService interface {
	// Upload uploads a file to storage
	Upload(ctx context.Context, req *UploadRequest) (*FileInfo, error)
	
	// Download generates a presigned URL for downloading
	Download(ctx context.Context, req *DownloadRequest) (string, error)
	
	// Delete removes a file from storage
	Delete(ctx context.Context, key string) error
	
	// GetFileInfo retrieves file information
	GetFileInfo(ctx context.Context, key string) (*FileInfo, error)
	
	// ListFiles lists files in a folder
	ListFiles(ctx context.Context, folder string, limit int) ([]*FileInfo, error)
	
	// GeneratePresignedUploadURL generates a presigned URL for direct upload
	GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error)
	
	// GeneratePresignedDownloadURL generates a presigned URL for direct download
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Provider        string `json:"provider"`         // aws, gcp, azure, local
	Region          string `json:"region"`           // AWS region
	Bucket          string `json:"bucket"`           // S3 bucket name
	AccessKeyID     string `json:"access_key_id"`    // AWS access key
	SecretAccessKey string `json:"secret_access_key"` // AWS secret key
	Endpoint        string `json:"endpoint"`         // Custom endpoint (for MinIO, etc.)
	UseSSL          bool   `json:"use_ssl"`          // Use HTTPS
	BaseURL         string `json:"base_url"`         // Base URL for public access
}

// FileType represents supported file types
type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeVideo    FileType = "video"
	FileTypeDocument FileType = "document"
	FileTypeOther    FileType = "other"
)

// GetFileType determines file type from content type
func GetFileType(contentType string) FileType {
	switch {
	case contentType == "image/jpeg" || contentType == "image/png" || contentType == "image/gif" || contentType == "image/webp":
		return FileTypeImage
	case contentType == "video/mp4" || contentType == "video/avi" || contentType == "video/mov":
		return FileTypeVideo
	case contentType == "application/pdf" || contentType == "application/msword":
		return FileTypeDocument
	default:
		return FileTypeOther
	}
}

// ValidateImageFile validates if the file is a supported image
func ValidateImageFile(contentType string, size int64) error {
	// Check content type
	if GetFileType(contentType) != FileTypeImage {
		return ErrUnsupportedFileType
	}
	
	// Check file size (max 10MB for images)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if size > maxSize {
		return ErrFileTooLarge
	}
	
	return nil
}

// Storage errors
var (
	ErrFileNotFound        = errors.New("file not found")
	ErrUnsupportedFileType = errors.New("unsupported file type")
	ErrFileTooLarge        = errors.New("file size exceeds limit")
	ErrUploadFailed        = errors.New("failed to upload file")
	ErrDownloadFailed      = errors.New("failed to download file")
	ErrDeleteFailed        = errors.New("failed to delete file")
)
