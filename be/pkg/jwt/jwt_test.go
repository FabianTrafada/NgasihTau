package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestManager() *Manager {
	return NewManager(Config{
		Secret:             "test-secret-key-32-chars-long!!",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	})
}

func TestGenerateAccessToken_Success(t *testing.T) {
	m := newTestManager()
	userID := uuid.New()

	token, err := m.GenerateAccessToken(userID, "student")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	m := newTestManager()
	userID := uuid.New()

	token, err := m.GenerateRefreshToken(userID, "student")
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestGenerateTokenPair_Success(t *testing.T) {
	m := newTestManager()
	userID := uuid.New()

	accessToken, refreshToken, err := m.GenerateTokenPair(userID, "teacher")
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}
	if accessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}
	if accessToken == refreshToken {
		t.Error("Access and refresh tokens should be different")
	}
}

func TestValidateAccessToken_Success(t *testing.T) {
	m := newTestManager()
	userID := uuid.New()
	role := "student"

	token, _ := m.GenerateAccessToken(userID, role)
	claims, err := m.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Errorf("Expected role %s, got %s", role, claims.Role)
	}
	if claims.TokenType != AccessToken {
		t.Errorf("Expected token type %s, got %s", AccessToken, claims.TokenType)
	}
}

func TestValidateRefreshToken_Success(t *testing.T) {
	m := newTestManager()
	userID := uuid.New()

	token, _ := m.GenerateRefreshToken(userID, "student")
	claims, err := m.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}
	if claims.TokenType != RefreshToken {
		t.Errorf("Expected token type %s, got %s", RefreshToken, claims.TokenType)
	}
}

func TestValidateAccessToken_WrongType(t *testing.T) {
	m := newTestManager()
	userID := uuid.New()

	// Generate refresh token but validate as access token
	token, _ := m.GenerateRefreshToken(userID, "student")
	_, err := m.ValidateAccessToken(token)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateRefreshToken_WrongType(t *testing.T) {
	m := newTestManager()
	userID := uuid.New()

	// Generate access token but validate as refresh token
	token, _ := m.GenerateAccessToken(userID, "student")
	_, err := m.ValidateRefreshToken(token)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	m := newTestManager()

	_, err := m.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	m1 := newTestManager()
	m2 := NewManager(Config{
		Secret:             "different-secret-key-32-chars!!",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
	})

	token, _ := m1.GenerateAccessToken(uuid.New(), "student")
	_, err := m2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for wrong secret")
	}
}

func TestGetAccessTokenExpiry(t *testing.T) {
	m := newTestManager()
	if m.GetAccessTokenExpiry() != 15*time.Minute {
		t.Errorf("Expected 15 minutes, got %v", m.GetAccessTokenExpiry())
	}
}

func TestGetRefreshTokenExpiry(t *testing.T) {
	m := newTestManager()
	if m.GetRefreshTokenExpiry() != 7*24*time.Hour {
		t.Errorf("Expected 7 days, got %v", m.GetRefreshTokenExpiry())
	}
}
