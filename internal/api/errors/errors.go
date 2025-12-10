package errors

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     string      `json:"error"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(errorType, message string, details interface{}) *ErrorResponse {
	return &ErrorResponse{
		Error:     errorType,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// RespondWithError sends an error response to the client
func RespondWithError(c *gin.Context, statusCode int, errorType, message string, details interface{}) {
	c.JSON(statusCode, NewErrorResponse(errorType, message, details))
}

// Common error types
const (
	ErrBadRequest    = "bad_request"
	ErrUnauthorized  = "unauthorized"
	ErrNotFound      = "not_found"
	ErrConflict      = "conflict"
	ErrInternalError = "internal_error"
)

// Common error helpers

// BadRequest returns a 400 Bad Request error
func BadRequest(c *gin.Context, message string, details interface{}) {
	RespondWithError(c, http.StatusBadRequest, ErrBadRequest, message, details)
}

// Unauthorized returns a 401 Unauthorized error
func Unauthorized(c *gin.Context, message string) {
	RespondWithError(c, http.StatusUnauthorized, ErrUnauthorized, message, nil)
}

// NotFound returns a 404 Not Found error
func NotFound(c *gin.Context, message string) {
	RespondWithError(c, http.StatusNotFound, ErrNotFound, message, nil)
}

// Conflict returns a 409 Conflict error
func Conflict(c *gin.Context, message string, details interface{}) {
	RespondWithError(c, http.StatusConflict, ErrConflict, message, details)
}

// InternalError returns a 500 Internal Server Error
func InternalError(c *gin.Context, message string, details interface{}) {
	RespondWithError(c, http.StatusInternalServerError, ErrInternalError, message, details)
}
