package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// LogRequestStart logs the start of a request with common fields
func LogRequestStart(logger *zap.Logger, c *fiber.Ctx, operation string) {
	logger.Info(operation+" attempt started",
		zap.String("trace_id", getTraceID(c)),
		zap.String("ip", c.IP()),
		zap.String("user_agent", c.Get("User-Agent")),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("content_type", c.Get("Content-Type")))
}

// LogRequestParsed logs successful request parsing with trace ID
func LogRequestParsed(logger *zap.Logger, c *fiber.Ctx, operation string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("trace_id", getTraceID(c)),
		zap.String("operation", operation),
	}, fields...)
	logger.Info("Request parsed successfully", allFields...)
}

// LogValidationError logs validation errors with trace ID
func LogValidationError(logger *zap.Logger, c *fiber.Ctx, err error, operation string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("trace_id", getTraceID(c)),
		zap.Error(err),
		zap.String("operation", operation),
	}, fields...)
	logger.Error("Request validation failed", allFields...)
}

// LogServiceCall logs when calling service layer with trace ID
func LogServiceCall(logger *zap.Logger, c *fiber.Ctx, operation string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("trace_id", getTraceID(c)),
		zap.String("operation", operation),
	}, fields...)
	logger.Info("Calling service layer", allFields...)
}

// LogServiceError logs service layer errors with trace ID
func LogServiceError(logger *zap.Logger, c *fiber.Ctx, err error, operation string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("trace_id", getTraceID(c)),
		zap.Error(err),
		zap.String("error_message", err.Error()),
		zap.String("operation", operation),
	}, fields...)
	logger.Error("Service layer error", allFields...)
}

// LogServiceSuccess logs successful service operations with trace ID
func LogServiceSuccess(logger *zap.Logger, c *fiber.Ctx, operation string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("trace_id", getTraceID(c)),
		zap.String("operation", operation),
	}, fields...)
	logger.Info("Service operation successful", allFields...)
}

// LogParsingError logs request parsing errors
func LogParsingError(logger *zap.Logger, err error, c *fiber.Ctx, operation string) {
	logger.Error("Failed to parse request body",
		zap.String("trace_id", getTraceID(c)),
		zap.Error(err),
		zap.String("operation", operation),
		zap.String("content_type", c.Get("Content-Type")),
		zap.String("body_preview", truncateString(string(c.Body()), 200)))
}

// LogRequestError logs general request errors
func LogRequestError(logger *zap.Logger, c *fiber.Ctx, message string, err error) {
	logger.Error(message,
		zap.String("trace_id", getTraceID(c)),
		zap.Error(err),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("ip", c.IP()),
		zap.String("content_type", c.Get("Content-Type")))
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SafeTokenLog returns a safe version of token for logging (first 8 chars)
func SafeTokenLog(token string) string {
	if len(token) == 0 {
		return "empty"
	}
	if len(token) <= 8 {
		return token
	}
	return token[:8] + "..."
}
