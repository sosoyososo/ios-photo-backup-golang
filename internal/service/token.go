package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ios-photo-backup/photo-backup-server/internal/models"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
)

// TokenService handles token management operations
type TokenService struct {
	tokenRepo *repository.TokenRepository
	jwtSecret string
}

// NewTokenService creates a new TokenService
func NewTokenService(tokenRepo *repository.TokenRepository, jwtSecret string) *TokenService {
	return &TokenService{
		tokenRepo: tokenRepo,
		jwtSecret: jwtSecret,
	}
}

// RefreshRequest represents a refresh token request
type RefreshRequest struct{}

// RefreshResponse represents a refresh token response
type RefreshResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Refresh refreshes a JWT token
func (s *TokenService) Refresh(tokenString string) (*RefreshResponse, error) {
	// Parse and validate the token
	claims, err := s.validateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Extract user ID from claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token exists in database and is not expired
	existingToken, err := s.tokenRepo.FindValidToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to find token: %w", err)
	}
	if existingToken == nil {
		return nil, fmt.Errorf("token not found or expired")
	}

	// Delete old token from database
	if err := s.tokenRepo.DeleteByTokenValue(tokenString); err != nil {
		return nil, fmt.Errorf("failed to delete old token: %w", err)
	}

	// Generate new token
	newClaims := jwt.MapClaims{
		"user_id":  userID,
		"username": claims["username"],
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := newToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	// Save new token to database
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	token := &models.Token{
		UserID:     uint(userID),
		TokenValue: newTokenString,
		CreatedAt:  time.Now(),
		ExpiresAt:  expiresAt,
	}

	if err := s.tokenRepo.Create(token); err != nil {
		return nil, fmt.Errorf("failed to save new token: %w", err)
	}

	return &RefreshResponse{
		Token:     newTokenString,
		ExpiresAt: expiresAt,
	}, nil
}

// validateToken validates a JWT token
func (s *TokenService) validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ValidateToken validates a JWT token (public method for middleware)
func (s *TokenService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	return s.validateToken(tokenString)
}

// StatusResponse represents a status check response
type StatusResponse struct {
	Status   string `json:"status"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

// Status checks authentication status
func (s *TokenService) Status(tokenString string) (*StatusResponse, error) {
	// Validate token
	claims, err := s.validateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if token exists in database and is not expired
	validToken, err := s.tokenRepo.FindValidToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to find token: %w", err)
	}
	if validToken == nil {
		return nil, fmt.Errorf("token not found or expired")
	}

	// Extract user info from claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &StatusResponse{
		Status:   "online",
		UserID:   uint(userID),
		Username: username,
	}, nil
}
