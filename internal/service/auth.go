package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ios-photo-backup/photo-backup-server/internal/config"
	"github.com/ios-photo-backup/photo-backup-server/internal/models"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
)

// AuthService handles authentication logic
type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	jwtSecret string
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtSecret: jwtSecret,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	// Find user by username
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := config.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	tokenString, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Save token to database
	token := &models.Token{
		UserID:    user.ID,
		TokenValue: tokenString,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	if err := s.tokenRepo.Create(token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return &LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
	}, nil
}

// generateToken generates a JWT token for a user
func (s *AuthService) generateToken(user *models.User) (string, time.Time, error) {
	// Token expires in 7 days
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create token
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}
