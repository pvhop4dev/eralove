package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eralove/eralove-backend/internal/domain"
	"go.uber.org/zap"
)

// LocalStorage implements StorageService using local filesystem
type LocalStorage struct {
	basePath string
	baseURL  string
	logger   *zap.Logger
}

// NewLocalStorage creates a new local storage service
func NewLocalStorage(basePath, baseURL string, logger *zap.Logger) (domain.StorageService, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  strings.TrimRight(baseURL, "/"),
		logger:   logger,
	}, nil
}

// Upload uploads a file to local storage
func (l *LocalStorage) Upload(ctx context.Context, req *domain.UploadRequest) (*domain.FileInfo, error) {
	// Generate unique key
	key := l.generateKey(req.Folder, req.UserID, req.Filename)
	filePath := filepath.Join(l.basePath, key)

	l.logger.Info("Starting local upload",
		zap.String("key", key),
		zap.String("path", filePath),
		zap.String("filename", req.Filename))

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		l.logger.Error("Failed to create directory", zap.Error(err), zap.String("dir", dir))
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		l.logger.Error("Failed to create file", zap.Error(err), zap.String("path", filePath))
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content
	written, err := io.Copy(file, req.File)
	if err != nil {
		l.logger.Error("Failed to write file", zap.Error(err), zap.String("path", filePath))
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Generate public URL
	url := l.generatePublicURL(key)

	fileInfo := &domain.FileInfo{
		Key:         key,
		URL:         url,
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Size:        written,
		UploadedAt:  time.Now(),
		Bucket:      "local",
	}

	l.logger.Info("Local upload successful",
		zap.String("key", key),
		zap.String("url", url),
		zap.Int64("size", written))

	return fileInfo, nil
}

// Download generates a URL for downloading (for local storage, this is just the public URL)
func (l *LocalStorage) Download(ctx context.Context, req *domain.DownloadRequest) (string, error) {
	return l.generatePublicURL(req.Key), nil
}

// Delete removes a file from local storage
func (l *LocalStorage) Delete(ctx context.Context, key string) error {
	filePath := filepath.Join(l.basePath, key)
	
	l.logger.Info("Deleting local file", zap.String("key", key), zap.String("path", filePath))

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return domain.ErrFileNotFound
		}
		l.logger.Error("Failed to delete local file", zap.Error(err), zap.String("path", filePath))
		return fmt.Errorf("failed to delete file: %w", err)
	}

	l.logger.Info("Local file deleted successfully", zap.String("key", key))
	return nil
}

// GetFileInfo retrieves file information from local storage
func (l *LocalStorage) GetFileInfo(ctx context.Context, key string) (*domain.FileInfo, error) {
	filePath := filepath.Join(l.basePath, key)
	
	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrFileNotFound
		}
		l.logger.Error("Failed to get file info", zap.Error(err), zap.String("path", filePath))
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	filename := filepath.Base(key)
	
	fileInfo := &domain.FileInfo{
		Key:        key,
		URL:        l.generatePublicURL(key),
		Filename:   filename,
		Size:       stat.Size(),
		UploadedAt: stat.ModTime(),
		Bucket:     "local",
	}

	return fileInfo, nil
}

// ListFiles lists files in a folder
func (l *LocalStorage) ListFiles(ctx context.Context, folder string, limit int) ([]*domain.FileInfo, error) {
	folderPath := filepath.Join(l.basePath, folder)
	
	var files []*domain.FileInfo
	count := 0

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || count >= limit {
			return nil
		}

		// Get relative path from base
		relPath, err := filepath.Rel(l.basePath, path)
		if err != nil {
			return err
		}

		// Convert to forward slashes for consistency
		key := filepath.ToSlash(relPath)

		fileInfo := &domain.FileInfo{
			Key:        key,
			URL:        l.generatePublicURL(key),
			Filename:   info.Name(),
			Size:       info.Size(),
			UploadedAt: info.ModTime(),
			Bucket:     "local",
		}

		files = append(files, fileInfo)
		count++
		return nil
	})

	if err != nil {
		l.logger.Error("Failed to list local files", zap.Error(err), zap.String("folder", folder))
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// GeneratePresignedUploadURL for local storage, returns the direct upload endpoint
func (l *LocalStorage) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	// For local storage, we can return an upload endpoint
	// This would need to be implemented in the HTTP handler
	return fmt.Sprintf("%s/upload?key=%s", l.baseURL, key), nil
}

// GeneratePresignedDownloadURL for local storage, returns the public URL
func (l *LocalStorage) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return l.generatePublicURL(key), nil
}

// generateKey creates a unique key for the file
func (l *LocalStorage) generateKey(folder, userID, filename string) string {
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
	
	return filepath.ToSlash(filepath.Join(folder, userID, uniqueFilename))
}

// generatePublicURL creates a public URL for the file
func (l *LocalStorage) generatePublicURL(key string) string {
	return fmt.Sprintf("%s/files/%s", l.baseURL, key)
}
