package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// getUserIDFromContext extracts user ID from fiber context
func getUserIDFromContext(c *fiber.Ctx) primitive.ObjectID {
	userID := c.Locals("user_id").(primitive.ObjectID)
	return userID
}

// ErrorResponse represents an error response
// @Description Error response structure
type ErrorResponse struct {
	Code    int         `json:"code" example:"409001"`                                   // Unique error code
	Error   string      `json:"error" example:"validation_failed"`                       // Error type
	Message string      `json:"message" example:"The provided data is invalid"`          // Human-readable error message
	TraceID string      `json:"trace_id" example:"550e8400-e29b-41d4-a716-446655440000"` // Request trace ID
	Details interface{} `json:"details,omitempty"`                                       // Additional error details (optional)
}

// SuccessResponse represents a success response
// @Description Success response structure
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`                                  // Success status
	Data    interface{} `json:"data,omitempty"`                                          // Response data (optional)
	Message string      `json:"message" example:"Operation completed successfully"`      // Success message
	TraceID string      `json:"trace_id" example:"550e8400-e29b-41d4-a716-446655440000"` // Request trace ID
}

// getTraceID extracts trace ID from fiber context
func getTraceID(c *fiber.Ctx) string {
	traceID := c.Locals("requestid")
	if traceID == nil {
		return c.Get("X-Request-ID", "unknown")
	}
	return traceID.(string)
}

// getLoggerWithTrace returns a logger with trace_id field
func getLoggerWithTrace(logger *zap.Logger, c *fiber.Ctx) *zap.Logger {
	return logger.With(zap.String("trace_id", getTraceID(c)))
}
