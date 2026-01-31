package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/offline/domain"
)

// Extend MockLicenseRepository with missing methods
func (m *MockLicenseRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.License, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.License), args.Error(1)
}

func (m *MockLicenseRepository) UpdateExpiration(ctx context.Context, id uuid.UUID, expiresAt time.Time) error {
	args := m.Called(ctx, id, expiresAt)
	return args.Error(0)
}

func (m *MockLicenseRepository) RevokeByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) error {
	args := m.Called(ctx, userID, materialID)
	return args.Error(0)
}

// MockLicenseMaterialAccessChecker is a mock implementation of LicenseMaterialAccessChecker.
type MockLicenseMaterialAccessChecker struct {
	mock.Mock
}

func (m *MockLicenseMaterialAccessChecker) CheckAccess(ctx context.Context, userID, materialID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, materialID)
	return args.Bool(0), args.Error(1)
}

// Test helper to create a test license
func createTestLicense(userID, materialID, deviceID uuid.UUID) *domain.License {
	return &domain.License{
		ID:                 uuid.New(),
		UserID:             userID,
		MaterialID:         materialID,
		DeviceID:           deviceID,
		Status:             domain.LicenseStatusActive,
		ExpiresAt:          time.Now().Add(30 * 24 * time.Hour),
		OfflineGracePeriod: 72 * time.Hour,
		LastValidatedAt:    time.Now(),
		Nonce:              "test-nonce-1234567890abcdef1234567890abcdef",
		CreatedAt:          time.Now(),
	}
}

func TestIssueLicense_Success(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)
	mockAccessChecker := new(MockLicenseMaterialAccessChecker)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, mockAccessChecker, nil)

	userID := uuid.New()
	materialID := uuid.New()
	device := createTestDevice(userID)

	input := IssueLicenseInput{
		UserID:      userID,
		MaterialID:  materialID,
		DeviceID:    device.ID,
		Fingerprint: device.Fingerprint,
	}

	// Device exists and fingerprint matches
	mockDeviceRepo.On("FindById", ctx, device.ID).Return(device, nil)
	// User has access to material
	mockAccessChecker.On("CheckAccess", ctx, userID, materialID).Return(true, nil)
	// No existing license
	mockLicenseRepo.On("FindByUserAndMaterial", ctx, userID, materialID, device.ID).Return(nil, errors.NotFound("license", ""))
	// Create license
	mockLicenseRepo.On("Create", ctx, mock.AnythingOfType("*domain.License")).Return(nil)
	// Update device last used
	mockDeviceRepo.On("UpdateLastUsed", ctx, device.ID).Return(nil)

	license, err := service.IssueLicense(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, license)
	assert.Equal(t, userID, license.UserID)
	assert.Equal(t, materialID, license.MaterialID)
	assert.Equal(t, device.ID, license.DeviceID)
	assert.Equal(t, domain.LicenseStatusActive, license.Status)
	assert.NotEmpty(t, license.Nonce)
	assert.True(t, license.ExpiresAt.After(time.Now()))
	mockLicenseRepo.AssertExpectations(t)
	mockDeviceRepo.AssertExpectations(t)
	mockAccessChecker.AssertExpectations(t)
}


func TestIssueLicense_ExistingActiveLicense(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)
	mockAccessChecker := new(MockLicenseMaterialAccessChecker)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, mockAccessChecker, nil)

	userID := uuid.New()
	materialID := uuid.New()
	device := createTestDevice(userID)
	existingLicense := createTestLicense(userID, materialID, device.ID)

	input := IssueLicenseInput{
		UserID:      userID,
		MaterialID:  materialID,
		DeviceID:    device.ID,
		Fingerprint: device.Fingerprint,
	}

	mockDeviceRepo.On("FindById", ctx, device.ID).Return(device, nil)
	mockAccessChecker.On("CheckAccess", ctx, userID, materialID).Return(true, nil)
	mockLicenseRepo.On("FindByUserAndMaterial", ctx, userID, materialID, device.ID).Return(existingLicense, nil)
	mockDeviceRepo.On("UpdateLastUsed", ctx, device.ID).Return(nil)

	license, err := service.IssueLicense(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, license)
	assert.Equal(t, existingLicense.ID, license.ID)
	mockLicenseRepo.AssertExpectations(t)
	mockDeviceRepo.AssertExpectations(t)
}

func TestIssueLicense_NoMaterialAccess(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)
	mockAccessChecker := new(MockLicenseMaterialAccessChecker)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, mockAccessChecker, nil)

	userID := uuid.New()
	materialID := uuid.New()
	device := createTestDevice(userID)

	input := IssueLicenseInput{
		UserID:      userID,
		MaterialID:  materialID,
		DeviceID:    device.ID,
		Fingerprint: device.Fingerprint,
	}

	mockDeviceRepo.On("FindById", ctx, device.ID).Return(device, nil)
	mockAccessChecker.On("CheckAccess", ctx, userID, materialID).Return(false, nil)

	license, err := service.IssueLicense(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, license)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "MATERIAL_ACCESS_DENIED", appErr.Message)
	mockDeviceRepo.AssertExpectations(t)
	mockAccessChecker.AssertExpectations(t)
}

func TestIssueLicense_DeviceNotFound(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, nil, nil)

	userID := uuid.New()
	deviceID := uuid.New()

	input := IssueLicenseInput{
		UserID:      userID,
		MaterialID:  uuid.New(),
		DeviceID:    deviceID,
		Fingerprint: "some-fingerprint",
	}

	mockDeviceRepo.On("FindById", ctx, deviceID).Return(nil, errors.NotFound("device", ""))

	license, err := service.IssueLicense(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, license)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "DEVICE_NOT_FOUND", appErr.Message)
	mockDeviceRepo.AssertExpectations(t)
}

func TestIssueLicense_FingerprintMismatch(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, nil, nil)

	userID := uuid.New()
	device := createTestDevice(userID)

	input := IssueLicenseInput{
		UserID:      userID,
		MaterialID:  uuid.New(),
		DeviceID:    device.ID,
		Fingerprint: "wrong-fingerprint",
	}

	mockDeviceRepo.On("FindById", ctx, device.ID).Return(device, nil)

	license, err := service.IssueLicense(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, license)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "DEVICE_FINGERPRINT_MISMATCH", appErr.Message)
	mockDeviceRepo.AssertExpectations(t)
}

func TestValidateLicense_Success(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, nil, nil)

	userID := uuid.New()
	materialID := uuid.New()
	device := createTestDevice(userID)
	license := createTestLicense(userID, materialID, device.ID)

	input := ValidateLicenseInput{
		LicenseID:   license.ID,
		DeviceID:    device.ID,
		Fingerprint: device.Fingerprint,
		Nonce:       license.Nonce,
	}

	mockLicenseRepo.On("FindByID", ctx, license.ID).Return(license, nil)
	mockDeviceRepo.On("FindByID", ctx, device.ID).Return(device, nil)
	mockLicenseRepo.On("UpdateValidation", ctx, license.ID, mock.AnythingOfType("string")).Return(nil)
	mockDeviceRepo.On("UpdateLastUsed", ctx, device.ID).Return(nil)

	result, err := service.ValidateLicense(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, license.Nonce, result.Nonce) // Nonce should be updated
	mockLicenseRepo.AssertExpectations(t)
	mockDeviceRepo.AssertExpectations(t)
}

func TestValidateLicense_InvalidNonce(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, nil, nil)

	userID := uuid.New()
	materialID := uuid.New()
	device := createTestDevice(userID)
	license := createTestLicense(userID, materialID, device.ID)

	input := ValidateLicenseInput{
		LicenseID:   license.ID,
		DeviceID:    device.ID,
		Fingerprint: device.Fingerprint,
		Nonce:       "wrong-nonce",
	}

	mockLicenseRepo.On("FindByID", ctx, license.ID).Return(license, nil)
	mockDeviceRepo.On("FindByID", ctx, device.ID).Return(device, nil)

	result, err := service.ValidateLicense(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "INVALID_NONCE", appErr.Message)
	mockLicenseRepo.AssertExpectations(t)
	mockDeviceRepo.AssertExpectations(t)
}

func TestValidateLicense_Expired(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, nil, nil)

	userID := uuid.New()
	materialID := uuid.New()
	device := createTestDevice(userID)
	license := createTestLicense(userID, materialID, device.ID)
	license.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired

	input := ValidateLicenseInput{
		LicenseID:   license.ID,
		DeviceID:    device.ID,
		Fingerprint: device.Fingerprint,
		Nonce:       license.Nonce,
	}

	mockLicenseRepo.On("FindByID", ctx, license.ID).Return(license, nil)
	mockDeviceRepo.On("FindByID", ctx, device.ID).Return(device, nil)

	result, err := service.ValidateLicense(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "LICENSE_EXPIRED", appErr.Message)
	mockLicenseRepo.AssertExpectations(t)
	mockDeviceRepo.AssertExpectations(t)
}

func TestValidateLicense_Revoked(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, nil, nil)

	userID := uuid.New()
	materialID := uuid.New()
	device := createTestDevice(userID)
	license := createTestLicense(userID, materialID, device.ID)
	license.Status = domain.LicenseStatusRevoked
	revokedAt := time.Now()
	license.RevokedAt = &revokedAt

	input := ValidateLicenseInput{
		LicenseID:   license.ID,
		DeviceID:    device.ID,
		Fingerprint: device.Fingerprint,
		Nonce:       license.Nonce,
	}

	mockLicenseRepo.On("FindByID", ctx, license.ID).Return(license, nil)
	mockDeviceRepo.On("FindByID", ctx, device.ID).Return(device, nil)

	result, err := service.ValidateLicense(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "LICENSE_REVOKED", appErr.Message)
	mockLicenseRepo.AssertExpectations(t)
	mockDeviceRepo.AssertExpectations(t)
}


func TestRenewLicense_Success(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, nil, nil)

	userID := uuid.New()
	materialID := uuid.New()
	device := createTestDevice(userID)
	license := createTestLicense(userID, materialID, device.ID)
	originalExpiresAt := license.ExpiresAt

	input := RenewLicenseInput{
		LicenseID:   license.ID,
		DeviceID:    device.ID,
		Fingerprint: device.Fingerprint,
	}

	mockLicenseRepo.On("FindByID", ctx, license.ID).Return(license, nil)
	mockDeviceRepo.On("FindByID", ctx, device.ID).Return(device, nil)
	mockLicenseRepo.On("UpdateExpiration", ctx, license.ID, mock.AnythingOfType("time.Time")).Return(nil)
	mockLicenseRepo.On("UpdateValidation", ctx, license.ID, mock.AnythingOfType("string")).Return(nil)
	mockDeviceRepo.On("UpdateLastUsed", ctx, device.ID).Return(nil)

	result, err := service.RenewLicense(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.ExpiresAt.After(originalExpiresAt))
	mockLicenseRepo.AssertExpectations(t)
	mockDeviceRepo.AssertExpectations(t)
}

func TestRenewLicense_Revoked(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, nil, nil)

	userID := uuid.New()
	materialID := uuid.New()
	device := createTestDevice(userID)
	license := createTestLicense(userID, materialID, device.ID)
	license.Status = domain.LicenseStatusRevoked
	revokedAt := time.Now()
	license.RevokedAt = &revokedAt

	input := RenewLicenseInput{
		LicenseID:   license.ID,
		DeviceID:    device.ID,
		Fingerprint: device.Fingerprint,
	}

	mockLicenseRepo.On("FindByID", ctx, license.ID).Return(license, nil)
	mockDeviceRepo.On("FindByID", ctx, device.ID).Return(device, nil)

	result, err := service.RenewLicense(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "LICENSE_REVOKED", appErr.Message)
	mockLicenseRepo.AssertExpectations(t)
	mockDeviceRepo.AssertExpectations(t)
}

func TestRevokeLicense_Success(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)

	service := NewLicenseService(mockLicenseRepo, nil, nil, nil)

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()
	license := createTestLicense(userID, materialID, deviceID)

	mockLicenseRepo.On("FindByID", ctx, license.ID).Return(license, nil)
	mockLicenseRepo.On("Revoke", ctx, license.ID).Return(nil)

	err := service.RevokeLicense(ctx, license.ID)

	assert.NoError(t, err)
	mockLicenseRepo.AssertExpectations(t)
}

func TestRevokeLicense_AlreadyRevoked(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)

	service := NewLicenseService(mockLicenseRepo, nil, nil, nil)

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()
	license := createTestLicense(userID, materialID, deviceID)
	license.Status = domain.LicenseStatusRevoked
	revokedAt := time.Now()
	license.RevokedAt = &revokedAt

	mockLicenseRepo.On("FindByID", ctx, license.ID).Return(license, nil)

	err := service.RevokeLicense(ctx, license.ID)

	assert.NoError(t, err) // Idempotent operation
	mockLicenseRepo.AssertExpectations(t)
}

func TestRevokeLicense_NotFound(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)

	service := NewLicenseService(mockLicenseRepo, nil, nil, nil)

	licenseID := uuid.New()

	mockLicenseRepo.On("FindByID", ctx, licenseID).Return(nil, errors.NotFound("license", ""))

	err := service.RevokeLicense(ctx, licenseID)

	assert.Error(t, err)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "LICENSE_NOT_FOUND", appErr.Message)
	mockLicenseRepo.AssertExpectations(t)
}

func TestRevokeByMaterial_Success(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)

	service := NewLicenseService(mockLicenseRepo, nil, nil, nil)

	materialID := uuid.New()

	mockLicenseRepo.On("RevokeByMaterialID", ctx, materialID).Return(nil)

	err := service.RevokeByMaterial(ctx, materialID)

	assert.NoError(t, err)
	mockLicenseRepo.AssertExpectations(t)
}

func TestRevokeByDevice_Success(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)

	service := NewLicenseService(mockLicenseRepo, nil, nil, nil)

	deviceID := uuid.New()

	mockLicenseRepo.On("RevokeByDeviceID", ctx, deviceID).Return(nil)

	err := service.RevokeByDevice(ctx, deviceID)

	assert.NoError(t, err)
	mockLicenseRepo.AssertExpectations(t)
}

func TestGetLicense_Success(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)

	service := NewLicenseService(mockLicenseRepo, nil, nil, nil)

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()
	license := createTestLicense(userID, materialID, deviceID)

	mockLicenseRepo.On("FindByID", ctx, license.ID).Return(license, nil)

	result, err := service.GetLicense(ctx, license.ID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, license.ID, result.ID)
	mockLicenseRepo.AssertExpectations(t)
}

func TestGetLicensesByUser_Success(t *testing.T) {
	ctx := context.Background()
	mockLicenseRepo := new(MockLicenseRepository)

	service := NewLicenseService(mockLicenseRepo, nil, nil, nil)

	userID := uuid.New()
	licenses := []*domain.License{
		createTestLicense(userID, uuid.New(), uuid.New()),
		createTestLicense(userID, uuid.New(), uuid.New()),
	}

	mockLicenseRepo.On("FindActiveByUserID", ctx, userID).Return(licenses, nil)

	result, err := service.GetLicensesByUser(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockLicenseRepo.AssertExpectations(t)
}

func TestGenerateNonce(t *testing.T) {
	// Test that nonces are unique
	nonces := make(map[string]bool)
	for i := 0; i < 100; i++ {
		nonce, err := generateNonce()
		assert.NoError(t, err)
		assert.Len(t, nonce, domain.NonceHexLength)
		assert.False(t, nonces[nonce], "nonce should be unique")
		nonces[nonce] = true
	}
}
