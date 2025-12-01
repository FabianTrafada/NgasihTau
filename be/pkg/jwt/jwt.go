// Package jwt provides JWT token generation and validation for authentication.
// It supports access tokens (short-lived) and refresh tokens (long-lived) with
// configurable expiration times.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType represents the type of JWT token.
type TokenType string

const (
	// AccessToken is a short-lived token for API authentication.
	AccessToken TokenType = "access"
	// RefreshToken is a long-lived token for obtaining new access tokens.
	RefreshToken TokenType = "refresh"
)

// Common errors returned by the JWT package.
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrTokenNotYetValid = errors.New("token is not yet valid")
)

// Claims represents the custom JWT claims for NgasihTau.
// Implements jwt.Claims interface.
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// Config holds JWT configuration settings.
type Config struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

// Manager handles JWT token operations.
type Manager struct {
	config Config
}

// NewManager creates a new JWT manager with the given configuration.
func NewManager(config Config) *Manager {
	if config.Issuer == "" {
		config.Issuer = "ngasihtau"
	}
	return &Manager{config: config}
}


// GenerateAccessToken creates a new access token for the given user.
// Access tokens are short-lived (default 15 minutes) and used for API authentication.
func (m *Manager) GenerateAccessToken(userID uuid.UUID, role string) (string, error) {
	return m.generateToken(userID, role, AccessToken, m.config.AccessTokenExpiry)
}

// GenerateRefreshToken creates a new refresh token for the given user.
// Refresh tokens are long-lived (default 7 days) and used to obtain new access tokens.
func (m *Manager) GenerateRefreshToken(userID uuid.UUID, role string) (string, error) {
	return m.generateToken(userID, role, RefreshToken, m.config.RefreshTokenExpiry)
}

// GenerateTokenPair creates both access and refresh tokens for the given user.
// Returns accessToken, refreshToken, and any error.
func (m *Manager) GenerateTokenPair(userID uuid.UUID, role string) (string, string, error) {
	accessToken, err := m.GenerateAccessToken(userID, role)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := m.GenerateRefreshToken(userID, role)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// generateToken creates a JWT token with the specified parameters.
func (m *Manager) generateToken(userID uuid.UUID, role string, tokenType TokenType, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.Secret))
}

// ValidateToken validates a JWT token and returns the claims if valid.
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.Secret), nil
	})

	if err != nil {
		return nil, m.mapJWTError(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ValidateAccessToken validates an access token and returns the claims.
// Returns an error if the token is not an access token.
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the claims.
// Returns an error if the token is not a refresh token.
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// mapJWTError maps jwt-go errors to our custom errors.
func (m *Manager) mapJWTError(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return ErrExpiredToken
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return ErrTokenNotYetValid
	case errors.Is(err, jwt.ErrSignatureInvalid):
		return ErrInvalidSignature
	case errors.Is(err, jwt.ErrTokenMalformed):
		return ErrInvalidToken
	default:
		return fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
}

// GetAccessTokenExpiry returns the configured access token expiry duration.
func (m *Manager) GetAccessTokenExpiry() time.Duration {
	return m.config.AccessTokenExpiry
}

// GetRefreshTokenExpiry returns the configured refresh token expiry duration.
func (m *Manager) GetRefreshTokenExpiry() time.Duration {
	return m.config.RefreshTokenExpiry
}

// RefreshTokenExpiry returns the expiration time for a new refresh token.
func (m *Manager) RefreshTokenExpiry() time.Time {
	return time.Now().Add(m.config.RefreshTokenExpiry)
}

// AccessTokenExpiry returns the expiration time for a new access token.
func (m *Manager) AccessTokenExpiry() time.Time {
	return time.Now().Add(m.config.AccessTokenExpiry)
}
