package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/offline/application"
	"ngasihtau/internal/offline/domain"
)

// MockLicenseService is a mock implementation of LicenseService.
type MockLicenseService struct {
	mock.Mock
}

func (m *MockLicenseService) IssueLicense(ctx context.Context, input application.IssueLicenseInput) (*domain.License, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.License), args.Error(1)
}

func (m *MockLicenseService) ValidateLicense(ctx context.Context, input application.ValidateLicenseInput) (*domain.License, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.License), args.Error(1)
}

func (m *MockLicenseService) RenewLicense(ctx context.Context, input application.RenewLicenseInput) (*domain.License, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.License), args.Error(1)
}

func (m *MockLicenseService) RevokeLicense(ctx context.Context, licenseID uuid.UUID) error {
	args := m.Called(ctx, licenseID)
	return args.Error(0)
}

func (m *MockLicenseService) RevokeByMaterial(ctx context.Context, materialID uuid.UUID) error {
	args := m.Called(ctx, materialID)
	return args.Error(0)
}

func (m *MockLicenseService) RevokeByDevice(ctx context.Context, deviceID uuid.UUID) error {
	args := m.Called(ctx, deviceID)
	return args.Error(0)
}

func (m *MockLicenseService) GetLicense(ctx context.Context, licenseID uuid.UUID) (*domain.License, error) {
	args := m.Called(ctx, licenseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.License), args.Error(1)
}

func (m *MockLicenseService) GetLicensesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.License, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.License), args.Error(1)
}

// setupLicenseTestApp creates a test Fiber app with the license handler.
func setupLicenseTestApp(licenseService application.LicenseService) *fiber.App {
	app := fiber.New()
	handler := NewLicenseHandler(licenseService)

	// Add a middleware to set user ID for testing
	app.Use(func(c *fiber.Ctx) error {
		userIDStr := c.Get("X-Test-User-ID")
		if userIDStr != "" {
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				c.Locals(middleware.UserIDKey, userID)
			}
		}
		return c.Next()
	})

	// Register routes
	app.Post("/materials/:material_id/license", handler.IssueLicense)
	app.Post("/licenses/:license_id/validate", handler.ValidateLicense)
	app.Post("/licenses/:license_id/renew", handler.RenewLicense)

	return app
}

// createTestLicense creates a test license.
func createTestLicense(userID, materialID, deviceID uuid.UUID) *domain.License {
	now := time.Now()
	return &domain.License{
		ID:                 uuid.New(),
		UserID:             userID,
		MaterialID:         materialID,
		DeviceID:           deviceID,
		Status:             domain.LicenseStatusActive,
		ExpiresAt:          now.Add(30 * 24 * time.Hour),
		OfflineGracePeriod: 72 * time.Hour,
		LastValidatedAt:    now,
		Nonce:              "test-nonce-1234567890abcdef",
		CreatedAt:          now,
	}
}

// ============================================================================
// IssueLicense Tests
// ============================================================================

func TestLicenseHandler_IssueLicense(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()
	license := createTestLicense(userID, materialID, deviceID)

	mockService.On("IssueLicense", mock.Anything, mock.MatchedBy(func(input application.IssueLicenseInput) bool {
		return input.UserID == userID &&
			input.MaterialID == materialID &&
			input.DeviceID == deviceID &&
			input.Fingerprint == "test-fingerprint-12345678901234567890"
	})).Return(license, nil)

	reqBody := IssueLicenseRequest{
		DeviceID:    deviceID.String(),
		Fingerprint: "test-fingerprint-12345678901234567890",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/materials/"+materialID.String()+"/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_IssueLicense_ValidationError(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	materialID := uuid.New()

	// Missing required fields
	reqBody := IssueLicenseRequest{
		DeviceID:    "invalid-uuid",
		Fingerprint: "short", // Too short
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/materials/"+materialID.String()+"/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLicenseHandler_IssueLicense_InvalidMaterialID(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	deviceID := uuid.New()

	reqBody := IssueLicenseRequest{
		DeviceID:    deviceID.String(),
		Fingerprint: "test-fingerprint-12345678901234567890",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/materials/invalid-uuid/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLicenseHandler_IssueLicense_AccessDenied(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	mockService.On("IssueLicense", mock.Anything, mock.Anything).
		Return(nil, errors.New(errors.CodeForbidden, "MATERIAL_ACCESS_DENIED"))

	reqBody := IssueLicenseRequest{
		DeviceID:    deviceID.String(),
		Fingerprint: "test-fingerprint-12345678901234567890",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/materials/"+materialID.String()+"/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_IssueLicense_Unauthorized(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	materialID := uuid.New()
	deviceID := uuid.New()

	reqBody := IssueLicenseRequest{
		DeviceID:    deviceID.String(),
		Fingerprint: "test-fingerprint-12345678901234567890",
	}
	body, _ := json.Marshal(reqBody)

	// No user ID header
	req := httptest.NewRequest("POST", "/materials/"+materialID.String()+"/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ============================================================================
// ValidateLicense Tests
// ============================================================================

func TestLicenseHandler_ValidateLicense(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()
	license := createTestLicense(userID, materialID, deviceID)

	mockService.On("ValidateLicense", mock.Anything, mock.MatchedBy(func(input application.ValidateLicenseInput) bool {
		return input.LicenseID == license.ID &&
			input.Fingerprint == "test-fingerprint-12345678901234567890" &&
			input.Nonce == "test-nonce-1234567890abcdef"
	})).Return(license, nil)

	reqBody := ValidateLicenseRequest{
		Fingerprint: "test-fingerprint-12345678901234567890",
		Nonce:       "test-nonce-1234567890abcdef",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/licenses/"+license.ID.String()+"/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_ValidateLicense_ValidationError(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	licenseID := uuid.New()

	// Missing required fields
	reqBody := ValidateLicenseRequest{
		Fingerprint: "short", // Too short
		Nonce:       "short", // Too short
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/licenses/"+licenseID.String()+"/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLicenseHandler_ValidateLicense_InvalidLicenseID(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()

	reqBody := ValidateLicenseRequest{
		Fingerprint: "test-fingerprint-12345678901234567890",
		Nonce:       "test-nonce-1234567890abcdef",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/licenses/invalid-uuid/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLicenseHandler_ValidateLicense_LicenseExpired(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	licenseID := uuid.New()

	mockService.On("ValidateLicense", mock.Anything, mock.Anything).
		Return(nil, errors.New(errors.CodeForbidden, "LICENSE_EXPIRED"))

	reqBody := ValidateLicenseRequest{
		Fingerprint: "test-fingerprint-12345678901234567890",
		Nonce:       "test-nonce-1234567890abcdef",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/licenses/"+licenseID.String()+"/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_ValidateLicense_InvalidNonce(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	licenseID := uuid.New()

	mockService.On("ValidateLicense", mock.Anything, mock.Anything).
		Return(nil, errors.New(errors.CodeForbidden, "INVALID_NONCE"))

	reqBody := ValidateLicenseRequest{
		Fingerprint: "test-fingerprint-12345678901234567890",
		Nonce:       "wrong-nonce-abcdefgh",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/licenses/"+licenseID.String()+"/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_ValidateLicense_Unauthorized(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	licenseID := uuid.New()

	reqBody := ValidateLicenseRequest{
		Fingerprint: "test-fingerprint-12345678901234567890",
		Nonce:       "test-nonce-1234567890abcdef",
	}
	body, _ := json.Marshal(reqBody)

	// No user ID header
	req := httptest.NewRequest("POST", "/licenses/"+licenseID.String()+"/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ============================================================================
// RenewLicense Tests
// ============================================================================

func TestLicenseHandler_RenewLicense(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()
	license := createTestLicense(userID, materialID, deviceID)

	mockService.On("RenewLicense", mock.Anything, mock.MatchedBy(func(input application.RenewLicenseInput) bool {
		return input.LicenseID == license.ID &&
			input.Fingerprint == "test-fingerprint-12345678901234567890"
	})).Return(license, nil)

	reqBody := struct {
		Fingerprint string `json:"fingerprint"`
	}{
		Fingerprint: "test-fingerprint-12345678901234567890",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/licenses/"+license.ID.String()+"/renew", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_RenewLicense_ValidationError(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	licenseID := uuid.New()

	// Missing required fingerprint
	reqBody := struct {
		Fingerprint string `json:"fingerprint"`
	}{
		Fingerprint: "short", // Too short
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/licenses/"+licenseID.String()+"/renew", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLicenseHandler_RenewLicense_InvalidLicenseID(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()

	reqBody := struct {
		Fingerprint string `json:"fingerprint"`
	}{
		Fingerprint: "test-fingerprint-12345678901234567890",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/licenses/invalid-uuid/renew", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLicenseHandler_RenewLicense_LicenseRevoked(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	licenseID := uuid.New()

	mockService.On("RenewLicense", mock.Anything, mock.Anything).
		Return(nil, errors.New(errors.CodeForbidden, "LICENSE_REVOKED"))

	reqBody := struct {
		Fingerprint string `json:"fingerprint"`
	}{
		Fingerprint: "test-fingerprint-12345678901234567890",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/licenses/"+licenseID.String()+"/renew", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_RenewLicense_NotFound(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	userID := uuid.New()
	licenseID := uuid.New()

	mockService.On("RenewLicense", mock.Anything, mock.Anything).
		Return(nil, errors.New(errors.CodeNotFound, "LICENSE_NOT_FOUND"))

	reqBody := struct {
		Fingerprint string `json:"fingerprint"`
	}{
		Fingerprint: "test-fingerprint-12345678901234567890",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/licenses/"+licenseID.String()+"/renew", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_RenewLicense_Unauthorized(t *testing.T) {
	mockService := new(MockLicenseService)
	app := setupLicenseTestApp(mockService)

	licenseID := uuid.New()

	reqBody := struct {
		Fingerprint string `json:"fingerprint"`
	}{
		Fingerprint: "test-fingerprint-12345678901234567890",
	}
	body, _ := json.Marshal(reqBody)

	// No user ID header
	req := httptest.NewRequest("POST", "/licenses/"+licenseID.String()+"/renew", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ============================================================================
// Response Conversion Tests
// ============================================================================

func TestToLicenseResponse(t *testing.T) {
	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()
	license := createTestLicense(userID, materialID, deviceID)

	resp := ToLicenseResponse(license)

	assert.NotNil(t, resp)
	assert.Equal(t, license.ID, resp.ID)
	assert.Equal(t, license.UserID, resp.UserID)
	assert.Equal(t, license.MaterialID, resp.MaterialID)
	assert.Equal(t, license.DeviceID, resp.DeviceID)
	assert.Equal(t, string(license.Status), resp.Status)
	assert.Equal(t, license.Nonce, resp.Nonce)
	assert.Equal(t, license.OfflineGracePeriod.String(), resp.OfflineGracePeriod)
}

func TestToLicenseResponse_Nil(t *testing.T) {
	resp := ToLicenseResponse(nil)
	assert.Nil(t, resp)
}
