package domain

import "fmt"

// ErrorCode represents a unique error code
// Format: HTTPCODE + 3 digits (e.g., 400001 = Bad Request + Invalid Credentials)
type ErrorCode int

const (
	// 400xxx - Bad Request Errors
	ErrCodeInvalidRequest      ErrorCode = 400001 // Invalid request body/format
	ErrCodeValidationFailed    ErrorCode = 400002 // Validation failed
	ErrCodeRequiredField       ErrorCode = 400003 // Required field missing
	ErrCodeInvalidFormat       ErrorCode = 400004 // Invalid data format
	ErrCodeWeakPassword        ErrorCode = 400005 // Password too weak
	ErrCodePasswordMismatch    ErrorCode = 400006 // Passwords don't match
	ErrCodeInvalidEmail        ErrorCode = 400007 // Invalid email format
	ErrCodeUnsupportedFileType ErrorCode = 400008 // Unsupported file type
	ErrCodeFileTooLarge        ErrorCode = 400009 // File size exceeds limit
	ErrCodeInvalidMatchRequest ErrorCode = 400010 // Invalid match request

	// 401xxx - Unauthorized Errors
	ErrCodeUnauthorized             ErrorCode = 401001 // Unauthorized access
	ErrCodeInvalidCredentials       ErrorCode = 401002 // Invalid email/password
	ErrCodeInvalidToken             ErrorCode = 401003 // Invalid or expired token
	ErrCodeTokenExpired             ErrorCode = 401004 // Token has expired
	ErrCodeInvalidVerificationToken ErrorCode = 401005 // Invalid verification token
	ErrCodeInvalidResetToken        ErrorCode = 401006 // Invalid reset token

	// 403xxx - Forbidden Errors
	ErrCodeForbidden        ErrorCode = 403001 // Access forbidden
	ErrCodeEmailNotVerified ErrorCode = 403002 // Email not verified

	// 404xxx - Not Found Errors
	ErrCodeNotFound             ErrorCode = 404001 // Resource not found
	ErrCodeUserNotFound         ErrorCode = 404002 // User not found
	ErrCodePhotoNotFound        ErrorCode = 404003 // Photo not found
	ErrCodeEventNotFound        ErrorCode = 404004 // Event not found
	ErrCodeMessageNotFound      ErrorCode = 404005 // Message not found
	ErrCodeMatchRequestNotFound ErrorCode = 404006 // Match request not found
	ErrCodeFileNotFound         ErrorCode = 404007 // File not found
	ErrCodeConversationNotFound ErrorCode = 404008 // Conversation not found

	// 409xxx - Conflict Errors
	ErrCodeUserAlreadyExists    ErrorCode = 409001 // User already exists
	ErrCodeEmailAlreadyVerified ErrorCode = 409002 // Email already verified
	ErrCodeMatchRequestExists   ErrorCode = 409003 // Match request already exists

	// 410xxx - Gone Errors
	ErrCodeMatchRequestExpired ErrorCode = 410001 // Match request expired

	// 500xxx - Internal Server Errors
	ErrCodeInternalError          ErrorCode = 500001 // Internal server error
	ErrCodeDatabaseError          ErrorCode = 500002 // Database error
	ErrCodeCacheError             ErrorCode = 500003 // Cache error
	ErrCodeFileUploadFailed       ErrorCode = 500004 // File upload failed
	ErrCodeFileDeleteFailed       ErrorCode = 500005 // File delete failed
	ErrCodePhotoUploadFailed      ErrorCode = 500006 // Photo upload failed
	ErrCodePhotoDeleteFailed      ErrorCode = 500007 // Photo delete failed
	ErrCodeEventCreateFailed      ErrorCode = 500008 // Event create failed
	ErrCodeEventUpdateFailed      ErrorCode = 500009 // Event update failed
	ErrCodeEventDeleteFailed      ErrorCode = 500010 // Event delete failed
	ErrCodeMessageSendFailed      ErrorCode = 500011 // Message send failed
	ErrCodeMessageDeleteFailed    ErrorCode = 500012 // Message delete failed
	ErrCodeMatchRequestFailed     ErrorCode = 500013 // Match request failed
	ErrCodeProfileUpdateFailed    ErrorCode = 500014 // Profile update failed
	ErrCodeAccountDeletionFailed  ErrorCode = 500015 // Account deletion failed
	ErrCodeOperationFailed        ErrorCode = 500016 // General operation failed
)

// AppError represents an application error with code and message
type AppError struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	StatusCode int         `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// Common error constructors
func ErrUserAlreadyExists(email string) *AppError {
	return NewAppError(
		ErrCodeUserAlreadyExists,
		fmt.Sprintf("User with email %s already exists", email),
		409,
	)
}

func ErrInvalidCredentials() *AppError {
	return NewAppError(
		ErrCodeInvalidCredentials,
		"Invalid email or password",
		401,
	)
}

func ErrUserNotFoundError() *AppError {
	return NewAppError(
		ErrCodeUserNotFound,
		"User not found",
		404,
	)
}

func ErrEmailNotVerifiedError() *AppError {
	return NewAppError(
		ErrCodeEmailNotVerified,
		"Please verify your email address first",
		403,
	)
}

func ErrInvalidTokenError() *AppError {
	return NewAppError(
		ErrCodeInvalidToken,
		"Invalid or expired token",
		401,
	)
}

func ErrTokenExpiredError() *AppError {
	return NewAppError(
		ErrCodeTokenExpired,
		"Token has expired",
		401,
	)
}

func ErrUnauthorizedError() *AppError {
	return NewAppError(
		ErrCodeUnauthorized,
		"Unauthorized access",
		401,
	)
}

func ErrValidationFailedError(details interface{}) *AppError {
	return NewAppError(
		ErrCodeValidationFailed,
		"Validation failed",
		400,
	).WithDetails(details)
}

func ErrInvalidRequestError(message string) *AppError {
	return NewAppError(
		ErrCodeInvalidRequest,
		message,
		400,
	)
}

func ErrInternalServerError() *AppError {
	return NewAppError(
		ErrCodeInternalError,
		"Internal server error",
		500,
	)
}

func ErrNotFoundError(resource string) *AppError {
	return NewAppError(
		ErrCodeNotFound,
		fmt.Sprintf("%s not found", resource),
		404,
	)
}

func ErrForbiddenError() *AppError {
	return NewAppError(
		ErrCodeForbidden,
		"Access forbidden",
		403,
	)
}

func ErrFileUploadFailedError(reason string) *AppError {
	return NewAppError(
		ErrCodeFileUploadFailed,
		fmt.Sprintf("File upload failed: %s", reason),
		500,
	)
}

func ErrUnsupportedFileTypeError(fileType string) *AppError {
	return NewAppError(
		ErrCodeUnsupportedFileType,
		fmt.Sprintf("Unsupported file type: %s", fileType),
		400,
	)
}

func ErrFileTooLargeError(maxSize int64) *AppError {
	return NewAppError(
		ErrCodeFileTooLarge,
		fmt.Sprintf("File size exceeds maximum allowed size of %d bytes", maxSize),
		400,
	)
}

// ErrUnauthorized is a simple error for unauthorized access
var ErrUnauthorized = ErrUnauthorizedError()
