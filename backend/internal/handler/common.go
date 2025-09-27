package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// getUserIDFromContext extracts user ID from fiber context
func getUserIDFromContext(c *fiber.Ctx) primitive.ObjectID {
	userID := c.Locals("user_id").(primitive.ObjectID)
	return userID
}

// ErrorResponse represents an error response
// @Description Error response structure
type ErrorResponse struct {
	Error   string   `json:"error" example:"validation_failed"`                    // Error code or type
	Message string   `json:"message" example:"The provided data is invalid"`       // Human-readable error message
	Details []string `json:"details,omitempty" example:"name is required,email is required"` // Additional error details (optional)
}

// SuccessResponse represents a success response
// @Description Success response structure
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`                           // Success status
	Data    interface{} `json:"data,omitempty"`                                   // Response data (optional)
	Message string      `json:"message" example:"Operation completed successfully"` // Success message
}
