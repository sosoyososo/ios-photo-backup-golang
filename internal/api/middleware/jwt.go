package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	apierrors "github.com/ios-photo-backup/photo-backup-server/internal/api/errors"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// Context keys for JWT claims
const (
	ClaimsKey  = "claims"
	UserIDKey  = "user_id"
	UsernameKey = "username"
)

// JWTMiddleware provides unified JWT authentication middleware
func JWTMiddleware(tokenService *service.TokenService, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			appLogger.Info("JWT missing Authorization header", logger.String("path", c.Request.URL.Path))
			apierrors.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			appLogger.Info("JWT invalid header format", logger.String("path", c.Request.URL.Path))
			apierrors.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			appLogger.Info("JWT empty token", logger.String("path", c.Request.URL.Path))
			apierrors.Unauthorized(c, "Token is required")
			c.Abort()
			return
		}

		// Validate token
		claims, err := tokenService.ValidateToken(token)
		if err != nil {
			appLogger.Warn("JWT validation failed",
				logger.String("path", c.Request.URL.Path),
				logger.String("error", err.Error()))
			apierrors.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// Extract user_id from claims
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			appLogger.Warn("JWT missing user_id in claims", logger.String("path", c.Request.URL.Path))
			apierrors.Unauthorized(c, "Invalid token claims: missing user_id")
			c.Abort()
			return
		}
		userID := uint(userIDFloat)

		// Store claims and user_id in context for handlers to use
		c.Set(ClaimsKey, claims)
		c.Set(UserIDKey, userID)

		// Optionally store username if present
		if username, ok := claims["username"].(string); ok {
			c.Set(UsernameKey, username)
		}

		c.Next()
	}
}

// GetUserID extracts user ID from Gin context (must be used after JWTMiddleware)
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

// GetClaims extracts all claims from Gin context (must be used after JWTMiddleware)
func GetClaims(c *gin.Context) (map[string]interface{}, bool) {
	claims, exists := c.Get(ClaimsKey)
	if !exists {
		return nil, false
	}
	data, ok := claims.(map[string]interface{})
	return data, ok
}

// GetUsername extracts username from Gin context (must be used after JWTMiddleware)
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(UsernameKey)
	if !exists {
		return "", false
	}
	name, ok := username.(string)
	return name, ok
}

// RequireUserID is a helper middleware that ensures user_id is present
// Use this in handlers if you need to verify user ID extraction succeeded
func RequireUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := GetUserID(c)
		if !ok {
			apierrors.Unauthorized(c, "User ID not found in context")
			c.Abort()
			return
		}
		c.Next()
	}
}
