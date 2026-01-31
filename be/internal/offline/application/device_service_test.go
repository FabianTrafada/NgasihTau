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

// MockDeviceRepository is a mock implementation of domain.DeviceRepository.
type MockDeviceRepository struct {
	mock.Mock
}

func (m *MockDeviceRepository) Create(ctx context.Context, device *domain.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Device), args.Error(1)
}

func (m *MockDeviceRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Device), args.Error(1)
}

func (m *MockDeviceRepository) FindByFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*domain.Device, error) {
	args := m.Called(ctx, userID, fingerprint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Device), args.Error(1)
}

func (m *MockDeviceRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockDeviceRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}


// MockLicenseRepository is a mock implementation of domain.LicenseRepository.
type MockLicenseRepository struct {
	mock.Mock
}

func (m *MockLicenseRepository) Create(ctx context.Context, license *domain.License) error {
	args := m.Called(ctx, license)
	return args.Error(0)
}

func (m *MockLicenseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.License, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.License), args.Error(1)
}

func (m *MockLicenseRepository) FindByUserAndMaterial(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*domain.License, error) {
	args := m.Called(ctx, userID, materialID, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.License), args.Error(1)
}

func (m *MockLicenseRepository) FindActiveByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*domain.License, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.License), args.Error(1)
}

func (m *MockLicenseRepository) UpdateValidation(ctx context.Context, id uuid.UUID, nonce string) error {
	args := m.Called(ctx, id, nonce)
	return args.Error(0)
}

func (m *MockLicenseRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLicenseRepository) RevokeByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	args := m.Called(ctx, deviceID)
	return args.Error(0)
}

func (m *MockLicenseRepository) RevokeByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	args := m.Called(ctx, materialID)
	return args.Error(0)
}

func (m *MockLicenseRepository) Renew(ctx context.Context, id uuid.UUID, newExpiresAt time.Time, newNonce string) error {
	args := m.Called(ctx, id, newExpiresAt, newNonce)
	return args.Error(0)
}

// MockCEKRepository is a mock implementation of domain.CEKRepository.
type MockCEKRepository struct {
	mock.Mock
}

func (m *MockCEKRepository) Create(ctx context.Context, cek *domain.ContentEncryptionKey) error {
	args := m.Called(ctx, cek)
	return args.Error(0)
}

func (m *MockCEKRepository) FindByComposite(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*domain.ContentEncryptionKey, error) {
	args := m.Called(ctx, userID, materialID, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ContentEncryptionKey), args.Error(1)
}

func (m *MockCEKRepository) DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	args := m.Called(ctx, deviceID)
	return args.Error(0)
}

// Test helper to create a test device
func createTestDevice(userID uuid.UUID) *domain.Device {
	return &domain.Device{
		ID:          uuid.New(),
		UserID:      userID,
		Fingerprint: "test-fingerprint-12345678901234567890",
		Name:        "Test Device",
		Platform:    domain.PlatformIOS,
		LastUsedAt:  time.Now(),
		CreatedAt:   time.Now(),
	}
}

func TestRegisterDevice_Success(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)
	mockLicenseRepo := new(MockLicenseRepository)
	mockCEKRepo := new(MockCEKRepository)

	service := NewDeviceService(mockDeviceRepo, mockLicenseRepo, mockCEKRepo, nil)

	userID := uuid.New()
	input := RegisterDeviceInput{
		UserID:      userID,
		Fingerprint: "new-fingerprint-12345678901234567890",
		Name:        "My iPhone",
		Platform:    domain.PlatformIOS,
	}

	// Device doesn't exist yet
	mockDeviceRepo.On("FindByFingerprint", ctx, userID, input.Fingerprint).Return(nil, errors.NotFound("device", ""))
	mockDeviceRepo.On("CountActiveByUserID", ctx, userID).Return(0, nil)
	mockDeviceRepo.On("Create", ctx, mock.AnythingOfType("*domain.Device")).Return(nil)

	device, err := service.RegisterDevice(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, userID, device.UserID)
	assert.Equal(t, input.Fingerprint, device.Fingerprint)
	assert.Equal(t, input.Name, device.Name)
	assert.Equal(t, input.Platform, device.Platform)
	mockDeviceRepo.AssertExpectations(t)
}

func TestRegisterDevice_ExistingDevice(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

	userID := uuid.New()
	existingDevice := createTestDevice(userID)
	input := RegisterDeviceInput{
		UserID:      userID,
		Fingerprint: existingDevice.Fingerprint,
		Name:        "My iPhone",
		Platform:    domain.PlatformIOS,
	}

	// Device already exists
	mockDeviceRepo.On("FindByFingerprint", ctx, userID, input.Fingerprint).Return(existingDevice, nil)
	mockDeviceRepo.On("UpdateLastUsed", ctx, existingDevice.ID).Return(nil)

	device, err := service.RegisterDevice(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, existingDevice.ID, device.ID)
	mockDeviceRepo.AssertExpectations(t)
}

func TestRegisterDevice_DeviceLimitExceeded(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

	userID := uuid.New()
	input := RegisterDeviceInput{
		UserID:      userID,
		Fingerprint: "new-fingerprint-12345678901234567890",
		Name:        "My iPhone",
		Platform:    domain.PlatformIOS,
	}

	// Device doesn't exist
	mockDeviceRepo.On("FindByFingerprint", ctx, userID, input.Fingerprint).Return(nil, errors.NotFound("device", ""))
	// User already has 5 devices
	mockDeviceRepo.On("CountActiveByUserID", ctx, userID).Return(5, nil)

	device, err := service.RegisterDevice(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, device)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "DEVICE_LIMIT_EXCEEDED", appErr.Message)
	mockDeviceRepo.AssertExpectations(t)
}

func TestRegisterDevice_InvalidPlatform(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

	input := RegisterDeviceInput{
		UserID:      uuid.New(),
		Fingerprint: "test-fingerprint-12345678901234567890",
		Name:        "My Device",
		Platform:    domain.Platform("invalid"),
	}

	device, err := service.RegisterDevice(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, device)
}


func TestListDevices_Success(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

	userID := uuid.New()
	devices := []*domain.Device{
		createTestDevice(userID),
		createTestDevice(userID),
	}

	mockDeviceRepo.On("FindByUserID", ctx, userID).Return(devices, nil)

	result, err := service.ListDevices(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockDeviceRepo.AssertExpectations(t)
}

func TestDeregisterDevice_Success(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)
	mockLicenseRepo := new(MockLicenseRepository)
	mockCEKRepo := new(MockCEKRepository)

	service := NewDeviceService(mockDeviceRepo, mockLicenseRepo, mockCEKRepo, nil)

	userID := uuid.New()
	device := createTestDevice(userID)

	mockDeviceRepo.On("FindByID", ctx, device.ID).Return(device, nil)
	mockLicenseRepo.On("RevokeByDeviceID", ctx, device.ID).Return(nil)
	mockCEKRepo.On("DeleteByDeviceID", ctx, device.ID).Return(nil)
	mockDeviceRepo.On("Revoke", ctx, device.ID).Return(nil)

	err := service.DeregisterDevice(ctx, userID, device.ID)

	assert.NoError(t, err)
	mockDeviceRepo.AssertExpectations(t)
	mockLicenseRepo.AssertExpectations(t)
	mockCEKRepo.AssertExpectations(t)
}

func TestDeregisterDevice_NotOwner(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	device := createTestDevice(ownerID)

	mockDeviceRepo.On("FindByID", ctx, device.ID).Return(device, nil)

	err := service.DeregisterDevice(ctx, otherUserID, device.ID)

	assert.Error(t, err)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, errors.CodeNotFound, appErr.Code)
	mockDeviceRepo.AssertExpectations(t)
}

func TestDeregisterDevice_AlreadyRevoked(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

	userID := uuid.New()
	device := createTestDevice(userID)
	revokedAt := time.Now()
	device.RevokedAt = &revokedAt

	mockDeviceRepo.On("FindByID", ctx, device.ID).Return(device, nil)

	err := service.DeregisterDevice(ctx, userID, device.ID)

	assert.Error(t, err)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, errors.CodeNotFound, appErr.Code)
	mockDeviceRepo.AssertExpectations(t)
}

func TestValidateDevice_Success(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

	userID := uuid.New()
	device := createTestDevice(userID)

	mockDeviceRepo.On("FindByFingerprint", ctx, userID, device.Fingerprint).Return(device, nil)

	result, err := service.ValidateDevice(ctx, userID, device.Fingerprint)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, device.ID, result.ID)
	mockDeviceRepo.AssertExpectations(t)
}

func TestValidateDevice_NotFound(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

	userID := uuid.New()
	fingerprint := "unknown-fingerprint"

	mockDeviceRepo.On("FindByFingerprint", ctx, userID, fingerprint).Return(nil, errors.NotFound("device", ""))

	result, err := service.ValidateDevice(ctx, userID, fingerprint)

	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "DEVICE_FINGERPRINT_MISMATCH", appErr.Message)
	mockDeviceRepo.AssertExpectations(t)
}

func TestValidateDevice_Revoked(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

	userID := uuid.New()
	device := createTestDevice(userID)
	revokedAt := time.Now()
	device.RevokedAt = &revokedAt

	mockDeviceRepo.On("FindByFingerprint", ctx, userID, device.Fingerprint).Return(device, nil)

	result, err := service.ValidateDevice(ctx, userID, device.Fingerprint)

	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, "DEVICE_NOT_FOUND", appErr.Message)
	mockDeviceRepo.AssertExpectations(t)
}

func TestUpdateLastUsed_Success(t *testing.T) {
	ctx := context.Background()
	mockDeviceRepo := new(MockDeviceRepository)

	service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

	deviceID := uuid.New()

	mockDeviceRepo.On("UpdateLastUsed", ctx, deviceID).Return(nil)

	err := service.UpdateLastUsed(ctx, deviceID)

	assert.NoError(t, err)
	mockDeviceRepo.AssertExpectations(t)
}
