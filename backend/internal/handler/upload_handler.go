package handler

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/i18n"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// UploadHandler handles file upload requests
type UploadHandler struct {
	storageService domain.StorageService
	i18n           *i18n.I18n
	logger         *zap.Logger
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(
	storageService domain.StorageService,
	i18n *i18n.I18n,
	logger *zap.Logger,
) *UploadHandler {
	return &UploadHandler{
		storageService: storageService,
		i18n:           i18n,
		logger:         logger,
	}
}

// UploadFileResponse represents the response after uploading a file
type UploadFileResponse struct {
	FilePath    string `json:"file_path"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	Message     string `json:"message"`
}

// UploadFile handles single file upload
// @Summary Upload a file
// @Description Upload a single file and get the file path
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param folder formData string false "Folder name (photos, avatars, documents)"
// @Security BearerAuth
// @Success 200 {object} UploadFileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /upload [post]
func (h *UploadHandler) UploadFile(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Upload file")

	userID := getUserIDFromContext(c)

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		LogRequestError(h.logger, c, "File upload failed", err)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "File is required",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	// Get optional folder parameter
	folder := c.FormValue("folder")
	if folder == "" {
		folder = "uploads"
	}

	// Validate file
	if err := h.validateFile(file); err != nil {
		LogRequestError(h.logger, c, "File validation failed", err)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid file",
			Message: err.Error(),
		})
	}

	LogRequestParsed(h.logger, c, "Upload file",
		zap.String("user_id", userID.Hex()),
		zap.String("filename", file.Filename),
		zap.Int64("size", file.Size),
		zap.String("folder", folder))

	// Generate unique filename
	timestamp := time.Now().Unix()
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%d%s", userID.Hex(), timestamp, ext)
	filePath := fmt.Sprintf("%s/%s", folder, filename)

	// Open file
	fileContent, err := file.Open()
	if err != nil {
		LogServiceError(h.logger, c, err, "Upload file", zap.String("user_id", userID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to read file",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}
	defer fileContent.Close()

	// Upload to storage
	LogServiceCall(h.logger, c, "Upload file", zap.String("file_path", filePath))
	
	uploadReq := &domain.UploadRequest{
		File:        fileContent,
		Filename:    file.Filename,
		ContentType: file.Header.Get("Content-Type"),
		Size:        file.Size,
		Folder:      folder,
		UserID:      userID.Hex(),
	}
	
	fileInfo, err := h.storageService.Upload(c.Context(), uploadReq)
	if err != nil {
		LogServiceError(h.logger, c, err, "Upload file", zap.String("file_path", filePath))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to upload file",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}
	
	url := fileInfo.URL

	LogServiceSuccess(h.logger, c, "Upload file",
		zap.String("user_id", userID.Hex()),
		zap.String("file_path", filePath),
		zap.String("url", url))

	return c.JSON(UploadFileResponse{
		FilePath:    filePath,
		FileName:    file.Filename,
		FileSize:    file.Size,
		ContentType: file.Header.Get("Content-Type"),
		URL:         url,
		Message:     "File uploaded successfully",
	})
}

// UploadMultipleFiles handles multiple file uploads
// @Summary Upload multiple files
// @Description Upload multiple files and get the file paths
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Files to upload" multiple
// @Param folder formData string false "Folder name (photos, avatars, documents)"
// @Security BearerAuth
// @Success 200 {object} map[string][]UploadFileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /upload/multiple [post]
func (h *UploadHandler) UploadMultipleFiles(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Upload multiple files")

	userID := getUserIDFromContext(c)

	// Get files from form
	form, err := c.MultipartForm()
	if err != nil {
		LogRequestError(h.logger, c, "Failed to parse multipart form", err)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid form data",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "No files provided",
			Message: "Please provide at least one file",
		})
	}

	folder := c.FormValue("folder")
	if folder == "" {
		folder = "uploads"
	}

	var responses []UploadFileResponse
	var errors []string

	for _, file := range files {
		// Validate file
		if err := h.validateFile(file); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", file.Filename, err.Error()))
			continue
		}

		// Generate unique filename
		timestamp := time.Now().UnixNano()
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%s_%d%s", userID.Hex(), timestamp, ext)
		filePath := fmt.Sprintf("%s/%s", folder, filename)

		// Open file
		fileContent, err := file.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to read", file.Filename))
			continue
		}

		// Upload to storage
		uploadReq := &domain.UploadRequest{
			File:        fileContent,
			Filename:    file.Filename,
			ContentType: file.Header.Get("Content-Type"),
			Size:        file.Size,
			Folder:      folder,
			UserID:      userID.Hex(),
		}
		
		fileInfo, err := h.storageService.Upload(c.Context(), uploadReq)
		fileContent.Close()

		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: upload failed", file.Filename))
			continue
		}
		
		url := fileInfo.URL

		responses = append(responses, UploadFileResponse{
			FilePath:    filePath,
			FileName:    file.Filename,
			FileSize:    file.Size,
			ContentType: file.Header.Get("Content-Type"),
			URL:         url,
			Message:     "File uploaded successfully",
		})
	}

	LogServiceSuccess(h.logger, c, "Upload multiple files",
		zap.String("user_id", userID.Hex()),
		zap.Int("uploaded", len(responses)),
		zap.Int("failed", len(errors)))

	result := map[string]interface{}{
		"files":   responses,
		"total":   len(files),
		"success": len(responses),
		"failed":  len(errors),
	}

	if len(errors) > 0 {
		result["errors"] = errors
	}

	return c.JSON(result)
}

// DeleteFile handles file deletion
// @Summary Delete a file
// @Description Delete a file by its path
// @Tags upload
// @Accept json
// @Produce json
// @Param request body map[string]string true "File path"
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /upload [delete]
func (h *UploadHandler) DeleteFile(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Delete file")

	userID := getUserIDFromContext(c)

	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Delete file")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	if req.FilePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "File path is required",
			Message: "Please provide file_path",
		})
	}

	LogServiceCall(h.logger, c, "Delete file",
		zap.String("user_id", userID.Hex()),
		zap.String("file_path", req.FilePath))

	if err := h.storageService.Delete(c.Context(), req.FilePath); err != nil {
		LogServiceError(h.logger, c, err, "Delete file", zap.String("file_path", req.FilePath))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to delete file",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, c, "Delete file",
		zap.String("user_id", userID.Hex()),
		zap.String("file_path", req.FilePath))

	return c.JSON(SuccessResponse{
		Message: "File deleted successfully",
	})
}

// validateFile validates the uploaded file
func (h *UploadHandler) validateFile(file *multipart.FileHeader) error {
	// Check file size (max 10MB)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if file.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of 10MB")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".mp4":  true,
		".mov":  true,
		".avi":  true,
	}

	if !allowedExts[ext] {
		return fmt.Errorf("file type %s is not allowed", ext)
	}

	return nil
}
