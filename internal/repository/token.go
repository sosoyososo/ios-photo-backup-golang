package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/ios-photo-backup/photo-backup-server/internal/models"
)

// TokenRepository provides CRUD operations for tokens
type TokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new TokenRepository
func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// Create creates a new token
func (r *TokenRepository) Create(token *models.Token) error {
	if err := r.db.Create(token).Error; err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}
	return nil
}

// FindByTokenValue finds a token by its value
func (r *TokenRepository) FindByTokenValue(tokenValue string) (*models.Token, error) {
	var token models.Token
	if err := r.db.Where("token_value = ?", tokenValue).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find token: %w", err)
	}
	return &token, nil
}

// FindValidToken finds a valid (non-expired) token
func (r *TokenRepository) FindValidToken(tokenValue string) (*models.Token, error) {
	var token models.Token
	if err := r.db.Where("token_value = ? AND expires_at > ?", tokenValue, time.Now()).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find valid token: %w", err)
	}
	return &token, nil
}

// DeleteByTokenValue deletes a token by its value
func (r *TokenRepository) DeleteByTokenValue(tokenValue string) error {
	if err := r.db.Where("token_value = ?", tokenValue).Delete(&models.Token{}).Error; err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

// DeleteExpiredTokens deletes all expired tokens
func (r *TokenRepository) DeleteExpiredTokens() error {
	if err := r.db.Where("expires_at <= ?", time.Now()).Delete(&models.Token{}).Error; err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}
	return nil
}

// GetTokensByUserID gets all tokens for a user
func (r *TokenRepository) GetTokensByUserID(userID uint) ([]models.Token, error) {
	var tokens []models.Token
	if err := r.db.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to get tokens for user: %w", err)
	}
	return tokens, nil
}
